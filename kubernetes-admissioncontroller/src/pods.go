package main

import (
	"encoding/json"
	"fmt"
	"github.com/kataras/golog"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strconv"
	"strings"
)

const (
	podEnvVarPatchTemplate string = `{"op":"add","path":"/spec/containers/%d/env","value":%s}`
)

var (
	ddEnvVar = corev1.EnvVar{
		Name: "DD_AGENT_HOST",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath: "status.hostIP",
			},
		},
	}
)

func mutatePods(ar v1beta1.AdmissionReview, client *kubernetes.Clientset) *v1beta1.AdmissionResponse {
	podResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	if ar.Request.Resource != podResource {
		golog.Errorf("expect resource to be %s", podResource)
		return nil
	}

	raw := ar.Request.Object.Raw
	var pod corev1.Pod
	if err := json.Unmarshal(raw, &pod); err != nil {
		golog.Error(err)
		return toAdmissionResponse(err)
	}
	reviewResponse := v1beta1.AdmissionResponse{}
	reviewResponse.Allowed = true
	if shouldPodBeMutated(&pod, client, ar.Request.Namespace) {
		var patchArr []string
		for k, v := range pod.Spec.Containers {
			patchEnv := v.Env
			patchEnv = append(patchEnv, ddEnvVar)
			jsonByteSlice, err := json.Marshal(patchEnv)
			if err != nil {
				golog.Error("error marshalling env vars into json: ", err)
			}
			patchArr = append(patchArr, fmt.Sprintf(podEnvVarPatchTemplate, k, string(jsonByteSlice)))
		}
		reviewResponse.Patch = []byte(fmt.Sprintf("[%v]", strings.Join(patchArr, ",")))
		pt := v1beta1.PatchTypeJSONPatch
		reviewResponse.PatchType = &pt
	}
	return &reviewResponse
}

//returns true if the datadog.com/apm-enabled label is set to true on one of the following:
// the pod
// the replicaset
// the deployment (it will automatically assume a replica set belongs to a deployment)
// the namespace
// it also respects inheritance, so if it should be enabled at the namespace level, but not at the deployment level it will not
func shouldPodBeMutated(pod *corev1.Pod, client *kubernetes.Clientset, ns string) bool {
	var boolArr []string
	nsResp, err := client.CoreV1().Namespaces().Get(ns, metav1.GetOptions{})
	if err != nil {
		golog.Error("error fetching namespace: ", err)
	}
	//order of ns -> deployment -> pod

	boolArr = appendOrDefer(boolArr, extractDatadogLabelValue(nsResp.Labels))
	//use v1beta1 for deployments to maintain backwards compatibility
	//this will also use a slight hack since currently a pod belongs to a replicaset which belongs to a deployment
	if len(pod.OwnerReferences) > 0 {
		golog.Info("Pod has an owner, attempting to fetch replicaset (if it belongs to one)")
		replSetResp, err := client.ExtensionsV1beta1().ReplicaSets(ns).Get(pod.OwnerReferences[0].Name, metav1.GetOptions{})
		if err != nil {
			golog.Error("error getting owner of pod: ", err)
		} else {
			if len(replSetResp.OwnerReferences) > 0 {
				deploymentResp, err := client.ExtensionsV1beta1().Deployments(ns).Get(replSetResp.OwnerReferences[0].Name, metav1.GetOptions{})
				if err != nil {
					golog.Error("error getting owner of replicaset (deployment): ", err)
				} else {
					boolArr = appendOrDefer(boolArr, extractDatadogLabelValue(deploymentResp.Labels))
				}
			}
			boolArr = appendOrDefer(boolArr, extractDatadogLabelValue(replSetResp.Labels))
		}
	}

	boolArr = appendOrDefer(boolArr, extractDatadogLabelValue(pod.Labels))

	//I'm intentionally using the fact that strconv.ParseBool produces an error in the presence of an empty string
	if len(boolArr) > 0 {
		flag, err := strconv.ParseBool(boolArr[len(boolArr) - 1])
		if err != nil {
			golog.Error(err)
		}
		return flag
	} else {
		return false
	}
}

//returns true if the datadog.com/apm-enabled label is in the map presented, or not present at all
func extractDatadogLabelValue(lbls map[string]string) string {
	if val, ok := lbls["datadog-apm-enabled"]; ok {
		return val
	} else {
		return ""
	}
}

func appendOrDefer(slice []string, elem string) []string {
	if elem != "" {
		return append(slice, elem)
	} else {
		return slice
	}
}
