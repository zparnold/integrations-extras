# (C) Datadog, Inc. 2018
# All rights reserved
# Licensed under a 3-clause BSD style license (see LICENSE)

from setuptools import setup
from codecs import open
from os import path

HERE = path.abspath(path.dirname(__file__))

# Get version info
ABOUT = {}
with open(path.join(HERE, "datadog_checks", "bind9", "__about__.py")) as f:
    exec(f.read(), ABOUT)

# Get the long description from the README file
with open(path.join(HERE, 'README.md'), encoding='utf-8') as f:
    long_description = f.read()


CHECKS_BASE_REQ = 'datadog-checks-base'

setup(
    name='datadog-bind9',
    version=ABOUT["__version__"],
    description='my_check collects my metrics.',
    long_description=long_description,
    keywords='datadog agent check',
    url='https://github.com/DataDog/integrations-extras',
    author='ashuvyas45',
    author_email='ashuvyas45@gmail.com',
    license='BSD',

    # See https://pypi.python.org/pypi?%3Aaction=list_classifiers
    classifiers=[
        'Development Status :: 5 - Production/Stable',
        'Intended Audience :: Developers',
        'Intended Audience :: System Administrators',
        'Topic :: System :: Monitoring',
        'License :: OSI Approved :: BSD License',
        'Programming Language :: Python :: 2.7',
        'Programming Language :: Python :: 3.7',
    ],

    packages=['datadog_checks.bind9'],

    # Run-time dependencies
    install_requires=[CHECKS_BASE_REQ],

    # Extra files to ship with the wheel package
    include_package_data=True,
)
