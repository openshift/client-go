#!/bin/bash
source "$(dirname "${BASH_SOURCE}")/../../hack/lib/init.sh"
trap os::test::junit::reconcile_output EXIT

# Cleanup cluster resources created by this test
(
  set +e
#  oc delete all,templates --all
  oc delete-project template-substitute
  oc delete-project prefix-template-substitute
  exit 0
) &>/dev/null


os::test::junit::declare_suite_start "cmd/newapp"
# This test validates the new-app command
os::cmd::expect_success_and_text 'oc new-app library/php mysql -o yaml' '3306'
os::cmd::expect_success_and_text 'oc new-app library/php mysql --dry-run' "Image \"library/php\" runs as the 'root' user which may not be permitted by your cluster administrator"
os::cmd::expect_failure 'oc new-app unknownhubimage -o yaml'
os::cmd::expect_failure_and_text 'oc new-app docker.io/node~https://github.com/openshift/nodejs-ex' 'the image match \"docker.io/node\" for source repository \"https://github.com/openshift/nodejs-ex\" does not appear to be a source-to-image builder.'
os::cmd::expect_failure_and_text 'oc new-app https://github.com/openshift/rails-ex' 'the image match \"ruby\" for source repository \"https://github.com/openshift/rails-ex\" does not appear to be a source-to-image builder.'
os::cmd::expect_success 'oc new-app https://github.com/openshift/rails-ex --strategy=source --dry-run'
# verify we can generate a Docker image based component "mongodb" directly
os::cmd::expect_success_and_text 'oc new-app mongo -o yaml' 'image:\s*mongo'
# the local image repository takes precedence over the Docker Hub "mysql" image
os::cmd::expect_success 'oc create -f examples/image-streams/image-streams-centos7.json'
os::cmd::try_until_success 'oc get imagestreamtags mysql:latest'
os::cmd::try_until_success 'oc get imagestreamtags mysql:5.5'
os::cmd::try_until_success 'oc get imagestreamtags mysql:5.6'
os::cmd::try_until_success 'oc get imagestreamtags mysql:5.7'
os::cmd::expect_success_and_not_text 'oc new-app mysql -o yaml' 'image:\s*mysql'
os::cmd::expect_success_and_not_text 'oc new-app mysql --dry-run' "runs as the 'root' user which may not be permitted by your cluster administrator"
# trigger and output should say 5.6
os::cmd::expect_success_and_text 'oc new-app mysql -o yaml' 'mysql:5.7'
os::cmd::expect_success_and_text 'oc new-app mysql --dry-run' 'tag "5.7" for "mysql"'
# test deployments are created with the boolean flag and printed in the UI
os::cmd::expect_success_and_text 'oc new-app mysql --dry-run --as-test' 'This image will be test deployed'
os::cmd::expect_success_and_text 'oc new-app mysql -o yaml --as-test' 'test: true'

# docker strategy with repo that has no dockerfile
os::cmd::expect_failure_and_text 'oc new-app https://github.com/openshift/nodejs-ex --strategy=docker' 'No Dockerfile was found'

# repo related error message validation
os::cmd::expect_failure_and_text 'oc new-app mysql-persisten mysql' 'mysql-persisten as a local directory'
os::cmd::expect_failure_and_text 'oc new-app mysql-persisten mysql' 'mysql as a local directory'
os::cmd::expect_failure_and_text 'oc new-app --strategy=docker https://192.30.253.113/openshift/ruby-hello-world.git' 'as a Git repository URL:  '
os::cmd::expect_failure_and_text 'oc new-app https://www.google.com/openshift/nodejs-e' 'as a Git repository URL:  '
os::cmd::expect_failure_and_text 'oc new-app https://examplegit.com/openshift/nodejs-e' 'as a Git repository URL:  '
os::cmd::expect_failure_and_text 'oc new-build --strategy=docker https://192.30.253.113/openshift/ruby-hello-world.git' 'as a Git repository URL:  '
os::cmd::expect_failure_and_text 'oc new-build https://www.google.com/openshift/nodejs-e' 'as a Git repository URL:  '
os::cmd::expect_failure_and_text 'oc new-build https://examplegit.com/openshift/nodejs-e' 'as a Git repository URL:  '

# setting source secret via the --source-secret flag
os::cmd::expect_success_and_text 'oc new-app https://github.com/openshift/cakephp-ex --source-secret=mysecret -o yaml' 'name: mysecret'
os::cmd::expect_success_and_text 'oc new-build https://github.com/openshift/cakephp-ex --source-secret=mynewsecret -o yaml' 'name: mynewsecret'
os::cmd::expect_success_and_text 'oc new-app -f examples/quickstarts/cakephp-mysql.json --source-secret=mysecret -o yaml' 'name: mysecret'
os::cmd::expect_success 'oc new-app https://github.com/openshift/cakephp-ex --source-secret=mysecret'
os::cmd::expect_success 'oc delete all --selector="label=cakephp-ex"'


# check label creation
os::cmd::try_until_success 'oc get imagestreamtags php:latest'
os::cmd::try_until_success 'oc get imagestreamtags php:5.5'
os::cmd::try_until_success 'oc get imagestreamtags php:5.6'
os::cmd::expect_success 'oc new-app php mysql -l no-source=php-mysql'
os::cmd::expect_success 'oc delete all -l no-source=php-mysql'
os::cmd::expect_success 'oc new-app php mysql'
os::cmd::expect_success 'oc delete all -l app=php'
os::cmd::expect_failure 'oc get dc/mysql'
os::cmd::expect_failure 'oc get dc/php'
os::cmd::expect_success_and_text 'oc new-app -f test/testdata/template-without-app-label.json -o yaml' 'app: ruby-helloworld-sample'

# check object namespace handling
# hardcoded values should be stripped
os::cmd::expect_success_and_not_text 'oc new-app -f test/testdata/template-with-namespaces.json -o jsonpath="{.items[?(@.metadata.name==\"stripped\")].metadata.namespace}"' 'STRIPPED'
# normal parameterized values should be substituted and retained
os::cmd::expect_success_and_text 'oc new-app -f test/testdata/template-with-namespaces.json -o jsonpath="{.items[?(@.metadata.name==\"route-edge-substituted\")].metadata.namespace}"' 'substituted'
os::cmd::expect_success_and_text 'oc new-app -f test/testdata/template-with-namespaces.json -o jsonpath="{.items[?(@.metadata.name==\"route-edge-prefix-substituted\")].metadata.namespace}"' 'prefix-substituted'
# non-string parameterized values should be stripped
os::cmd::expect_failure_and_text 'oc new-app -f test/testdata/template-with-namespaces.json -o jsonpath="{.items[?(@.metadata.name==\"route-edge-refstripped\")].metadata.namespace}"' 'namespace is not found'
os::cmd::expect_failure_and_text 'oc new-app -f test/testdata/template-with-namespaces.json -o jsonpath="{.items[?(@.metadata.name==\"route-edge-prefix-refstripped\")].metadata.namespace}"' 'namespace is not found'
# ensure --build-env environment variables get added to the buildconfig
os::cmd::expect_success_and_text 'oc new-app -f test/testdata/template-with-app-label.json --build-env FOO=bar -o yaml' 'FOO'
# ensure the objects can actually get created with a namespace specified
project=$(oc project -q)
os::cmd::expect_success 'oc new-project template-substitute'
os::cmd::expect_success 'oc new-project prefix-template-substitute'
os::cmd::expect_success 'oc project ${project}'
os::cmd::expect_success 'oc new-app -f test/testdata/template-with-namespaces.json -p SUBSTITUTED=template-substitute'
os::cmd::expect_success 'oc delete all -l app=ruby-helloworld-sample'

# ensure non-duplicate invalid label errors show up
os::cmd::expect_failure_and_text 'oc new-app nginx -l qwer1345%$$#=self' 'error: ImageStream "nginx" is invalid'
os::cmd::expect_failure_and_text 'oc new-app nginx -l qwer1345%$$#=self' 'DeploymentConfig "nginx" is invalid'
os::cmd::expect_failure_and_text 'oc new-app nginx -l qwer1345%$$#=self' 'Service "nginx" is invalid'

# check if we can create from a stored template
os::cmd::expect_success 'oc create -f examples/sample-app/application-template-stibuild.json'
os::cmd::expect_success 'oc get template ruby-helloworld-sample'
os::cmd::expect_success_and_text 'oc new-app ruby-helloworld-sample -o yaml' 'MYSQL_USER'
os::cmd::expect_success_and_text 'oc new-app ruby-helloworld-sample -o yaml' 'MYSQL_PASSWORD'
os::cmd::expect_success_and_text 'oc new-app ruby-helloworld-sample --param MYSQL_PASSWORD=hello -o yaml' 'hello'
os::cmd::expect_success_and_text  'oc new-app -e FOO=BAR -f examples/jenkins/jenkins-ephemeral-template.json -o jsonpath="{.items[?(@.kind==\"DeploymentConfig\")].spec.template.spec.containers[0].env[?(@.name==\"FOO\")].value}" ' '^BAR$'
os::cmd::expect_success_and_text  'oc new-app -e OPENSHIFT_ENABLE_OAUTH=false -f examples/jenkins/jenkins-ephemeral-template.json -o jsonpath="{.items[?(@.kind==\"DeploymentConfig\")].spec.template.spec.containers[0].env[?(@.name==\"OPENSHIFT_ENABLE_OAUTH\")].value}" ' 'false'

# check that multiple resource groups are printed with their respective external version
os::cmd::expect_success_and_text 'oc new-app -f test/testdata/template_multiple_resource_gvs.yaml -o yaml' 'apiVersion: apps/v1beta1'
# check that if an --output-version is requested on a list of varying resource kinds, an error is returned if
# at least one of the resource groups does not support the given version
os::cmd::expect_failure_and_text 'oc new-app -f test/testdata/template_multiple_resource_gvs.yaml -o yaml --output-version=v1' 'extensions.Deployment is not suitable for converting'
os::cmd::expect_failure_and_text 'oc new-app -f test/testdata/template_multiple_resource_gvs.yaml -o yaml --output-version=extensions/v1beta1' 'api.Secret is not suitable for converting'
os::cmd::expect_failure_and_not_text 'oc new-app -f test/testdata/template_multiple_resource_gvs.yaml -o yaml --output-version=apps/v1beta1' 'extensions.Deployment is not suitable for converting'

# check that an error is produced when using --context-dir with a template
os::cmd::expect_failure_and_text 'oc new-app -f examples/sample-app/application-template-stibuild.json --context-dir=example' '\-\-context-dir is not supported when using a template'

# check that values are not split on commas
os::cmd::expect_success_and_text 'oc new-app ruby-helloworld-sample --param MYSQL_DATABASE=hello,MYSQL_USER=fail -o yaml' 'value: hello,MYSQL_USER=fail'
# check that warning is printed when --param PARAM1=VAL1,PARAM2=VAL2 is used
os::cmd::expect_success_and_text 'oc new-app ruby-helloworld-sample --param MYSQL_DATABASE=hello,MYSQL_USER=fail -o yaml' 'no longer accepts comma-separated list'
# check that env vars are not split on commas
os::cmd::expect_success_and_text 'oc new-app php --env PASS=one,two=three -o yaml' 'value: one,two=three'
# check that warning is printed when --env PARAM1=VAL1,PARAM2=VAL2 is used
os::cmd::expect_success_and_text 'oc new-app php --env PASS=one,two=three -o yaml' 'no longer accepts comma-separated list'
# check that warning is not printed when --param/env doesn't contain two k-v pairs
os::cmd::expect_success_and_not_text 'oc new-app php --env DEBUG=disabled -o yaml' 'no longer accepts comma-separated list'
os::cmd::expect_success_and_not_text 'oc new-app php --env LEVELS=INFO,WARNING -o yaml' 'no longer accepts comma-separated list'
os::cmd::expect_success_and_not_text 'oc new-app ruby-helloworld-sample --param MYSQL_USER=mysql -o yaml' 'no longer accepts comma-separated list'
os::cmd::expect_success_and_not_text 'oc new-app ruby-helloworld-sample --param MYSQL_PASSWORD=com,ma -o yaml' 'no longer accepts comma-separated list'
# check that warning is not printed when env vars are passed positionally
os::cmd::expect_success_and_text 'oc new-app php PASS=one,two=three -o yaml' 'value: one,two=three'
os::cmd::expect_success_and_not_text 'oc new-app php PASS=one,two=three -o yaml' 'no longer accepts comma-separated list'

# check that we can populate template parameters from file
param_file="${OS_ROOT}/test/testdata/test-cmd-newapp-params.env"
os::cmd::expect_success_and_text "oc new-app ruby-helloworld-sample --param-file ${param_file} -o jsonpath='{.items[?(@.kind==\"DeploymentConfig\")].spec.template.spec.containers[0].env[?(@.name==\"MYSQL_DATABASE\")].value}'" 'thisisadatabase'
os::cmd::expect_success_and_text "oc new-app ruby-helloworld-sample --param-file ${param_file} --param MYSQL_DATABASE=otherdatabase -o jsonpath='{.items[?(@.kind==\"DeploymentConfig\")].spec.template.spec.containers[0].env[?(@.name==\"MYSQL_DATABASE\")].value}'" 'otherdatabase'
os::cmd::expect_success_and_text "oc new-app ruby-helloworld-sample --param-file ${param_file} --param MYSQL_DATABASE=otherdatabase -o yaml" 'ignoring value from file'
os::cmd::expect_success_and_text "cat ${param_file} | oc new-app ruby-helloworld-sample --param-file - -o jsonpath='{.items[?(@.kind==\"DeploymentConfig\")].spec.template.spec.containers[0].env[?(@.name==\"MYSQL_DATABASE\")].value}'" 'thisisadatabase'

os::cmd::expect_failure_and_text "oc new-app ruby-helloworld-sample --param-file does/not/exist" 'no such file or directory'
os::cmd::expect_failure_and_text "oc new-app ruby-helloworld-sample --param-file test/testdata"  'is a directory'
os::cmd::expect_success "oc new-app ruby-helloworld-sample --param-file /dev/null -o yaml"
os::cmd::expect_success "oc new-app ruby-helloworld-sample --param-file /dev/null --param-file ${param_file} -o yaml"
os::cmd::expect_failure_and_text "echo 'fo%(o=bar' | oc new-app ruby-helloworld-sample --param-file -" 'invalid parameter assignment'
os::cmd::expect_failure_and_text "echo 'S P A C E S=test' | oc new-app ruby-helloworld-sample --param-file -" 'invalid parameter assignment'

os::cmd::expect_failure_and_text 'oc new-app ruby-helloworld-sample --param ABSENT_PARAMETER=absent -o yaml' 'unexpected parameter name'
os::cmd::expect_success 'oc new-app ruby-helloworld-sample --param ABSENT_PARAMETER=absent -o yaml --ignore-unknown-parameters'

# check that we can set environment variables from env file
env_file="${OS_ROOT}/test/testdata/test-cmd-newapp-env.env"
os::cmd::expect_success_and_text "oc new-app php --env-file ${env_file} -o jsonpath='{.items[?(@.kind==\"DeploymentConfig\")].spec.template.spec.containers[0].env[?(@.name==\"SOME_VAR\")].value}'" 'envvarfromfile'
os::cmd::expect_success_and_text "oc new-app php --env-file ${env_file} --env SOME_VAR=fromcmdline -o jsonpath='{.items[?(@.kind==\"DeploymentConfig\")].spec.template.spec.containers[0].env[?(@.name==\"SOME_VAR\")].value}'" 'fromcmdline'
os::cmd::expect_success_and_text "oc new-app php --env-file ${env_file} --env SOME_VAR=fromcmdline -o yaml" 'ignoring value from file'
os::cmd::expect_success_and_text "cat ${env_file} | oc new-app php --env-file - -o jsonpath='{.items[?(@.kind==\"DeploymentConfig\")].spec.template.spec.containers[0].env[?(@.name==\"SOME_VAR\")].value}'" 'envvarfromfile'

os::cmd::expect_failure_and_text "oc new-app php --env-file does/not/exist" 'no such file or directory'
os::cmd::expect_failure_and_text "oc new-app php --env-file test/testdata"  'is a directory'
os::cmd::expect_success "oc new-app php --env-file /dev/null -o yaml"
os::cmd::expect_success "oc new-app php --env-file /dev/null --env-file ${env_file} -o yaml"
os::cmd::expect_failure_and_text "echo 'fo%(o=bar' | oc new-app php --env-file -" 'invalid parameter assignment'
os::cmd::expect_failure_and_text "echo 'S P A C E S=test' | oc new-app php --env-file -" 'invalid parameter assignment'

# new-build
# check that env vars are not split on commas and warning is printed where they previously have
os::cmd::expect_success_and_text 'oc new-build --binary php --env X=Y,Z=W -o yaml' 'value: Y,Z=W'
os::cmd::expect_success_and_text 'oc new-build --binary php --env X=Y,Z=W -o yaml' 'no longer accepts comma-separated list'
os::cmd::expect_success_and_text 'oc new-build --binary php --env X=Y,Z,W -o yaml' 'value: Y,Z,W'
os::cmd::expect_success_and_not_text 'oc new-build --binary php --env X=Y,Z,W -o yaml' 'no longer accepts comma-separated list'
os::cmd::expect_success_and_not_text 'oc new-build --binary php --env X=Y -o yaml' 'no longer accepts comma-separated list'

# new-build - load envvars from file
os::cmd::expect_success_and_text "oc new-build --binary php --env-file ${env_file} -o jsonpath='{.items[?(@.kind==\"BuildConfig\")].spec.strategy.sourceStrategy.env[?(@.name==\"SOME_VAR\")].value}'" 'envvarfromfile'
os::cmd::expect_success_and_text "oc new-build --binary php --env-file ${env_file} --env SOME_VAR=fromcmdline -o jsonpath='{.items[?(@.kind==\"BuildConfig\")].spec.strategy.sourceStrategy.env[?(@.name==\"SOME_VAR\")].value}'" 'fromcmdline'
os::cmd::expect_success_and_text "oc new-build --binary php --env-file ${env_file} --env SOME_VAR=fromcmdline -o yaml" 'ignoring value from file'
os::cmd::expect_success_and_text "cat ${env_file} | oc new-build --binary php --env-file ${env_file} -o jsonpath='{.items[?(@.kind==\"BuildConfig\")].spec.strategy.sourceStrategy.env[?(@.name==\"SOME_VAR\")].value}'" 'envvarfromfile'

os::cmd::expect_failure_and_text "oc new-build --binary php --env-file does/not/exist" 'no such file or directory'
os::cmd::expect_failure_and_text "oc new-build --binary php --env-file test/testdata"  'is a directory'
os::cmd::expect_success "oc new-build --binary php --env-file /dev/null -o yaml"
os::cmd::expect_success "oc new-build --binary php --env-file /dev/null --env-file ${env_file} -o yaml"
os::cmd::expect_failure_and_text "echo 'fo%(o=bar' | oc new-build --binary php --env-file -" 'invalid parameter assignment'
os::cmd::expect_failure_and_text "echo 'S P A C E S=test' | oc new-build --binary php --env-file -" 'invalid parameter assignment'

# check that we can set environment variables from build-env file
build_env_file="${OS_ROOT}/test/testdata/test-cmd-newapp-build-env.env"

os::cmd::expect_failure_and_text "oc new-app php --build-env-file does/not/exist" 'no such file or directory'
os::cmd::expect_failure_and_text "oc new-app php --build-env-file test/testdata"  'is a directory'
os::cmd::expect_success "oc new-app php --build-env-file /dev/null -o yaml"
os::cmd::expect_success "oc new-app php --build-env-file /dev/null --build-env-file ${build_env_file} -o yaml"
os::cmd::expect_failure_and_text "echo 'fo%(o=bar' | oc new-app php --build-env-file -" 'invalid parameter assignment'
os::cmd::expect_failure_and_text "echo 'S P A C E S=test' | oc new-app php --build-env-file -" 'invalid parameter assignment'

# new-build
# check that build env vars are not split on commas and warning is printed where they previously have
os::cmd::expect_success_and_text 'oc new-build --binary php --build-env X=Y,Z=W -o yaml' 'value: Y,Z=W'
os::cmd::expect_success_and_text 'oc new-build --binary php --build-env X=Y,Z=W -o yaml' 'no longer accepts comma-separated list'
os::cmd::expect_success_and_text 'oc new-build --binary php --build-env X=Y,Z,W -o yaml' 'value: Y,Z,W'
os::cmd::expect_success_and_not_text 'oc new-build --binary php --build-env X=Y,Z,W -o yaml' 'no longer accepts comma-separated list'
os::cmd::expect_success_and_not_text 'oc new-build --binary php --build-env X=Y -o yaml' 'no longer accepts comma-separated list'

# new-build - load build env vars from file
os::cmd::expect_success_and_text "oc new-build --binary php --build-env-file ${build_env_file} -o jsonpath='{.items[?(@.kind==\"BuildConfig\")].spec.strategy.sourceStrategy.env[?(@.name==\"SOME_VAR\")].value}'" 'buildenvvarfromfile'
os::cmd::expect_success_and_text "oc new-build --binary php --build-env-file ${build_env_file} --env SOME_VAR=fromcmdline -o jsonpath='{.items[?(@.kind==\"BuildConfig\")].spec.strategy.sourceStrategy.env[?(@.name==\"SOME_VAR\")].value}'" 'fromcmdline'
os::cmd::expect_success_and_text "oc new-build --binary php --build-env-file ${build_env_file} --env SOME_VAR=fromcmdline -o yaml" 'ignoring value from file'
os::cmd::expect_success_and_text "cat ${build_env_file} | oc new-build --binary php --build-env-file - -o jsonpath='{.items[?(@.kind==\"BuildConfig\")].spec.strategy.sourceStrategy.env[?(@.name==\"SOME_VAR\")].value}'" 'buildenvvarfromfile'

os::cmd::expect_failure_and_text "oc new-build --binary php --build-env-file does/not/exist" 'no such file or directory'
os::cmd::expect_failure_and_text "oc new-build --binary php --build-env-file test/testdata"  'is a directory'
os::cmd::expect_success "oc new-build --binary php --build-env-file /dev/null -o yaml"
os::cmd::expect_success "oc new-build --binary php --build-env-file /dev/null --env-file ${build_env_file} -o yaml"
os::cmd::expect_failure_and_text "echo 'fo%(o=bar' | oc new-build --binary php --build-env-file -" 'invalid parameter assignment'
os::cmd::expect_failure_and_text "echo 'S P A C E S=test' | oc new-build --binary php --build-env-file -" 'invalid parameter assignment'

# new-build - check that we can set build args in DockerStrategy
os::cmd::expect_success_and_text "oc new-build ${OS_ROOT}/test/testdata/build-arg-dockerfile --build-arg 'foo=bar' --to 'test' -o jsonpath='{.items[?(@.kind==\"BuildConfig\")].spec.strategy.dockerStrategy.buildArgs[?(@.name==\"foo\")].value}'" 'bar'

# check that we cannot set build args in a non-DockerStrategy build
os::cmd::expect_failure_and_text "oc new-build https://github.com/openshift/ruby-hello-world --strategy=source --build-arg 'foo=bar'" "error: Cannot use '--build-arg' without a Docker build"
os::cmd::expect_failure_and_text "oc new-build https://github.com/openshift/ruby-ex --build-arg 'foo=bar'" "error: Cannot use '--build-arg' without a Docker build"

#
# verify we can create from a template when some objects in the template declare an app label
# the app label will not be applied to any objects in the template.
os::cmd::expect_success_and_not_text 'oc new-app -f test/testdata/template-with-app-label.json -o yaml' 'app: ruby-helloworld-sample'
# verify the existing app label on an object is not overridden by new-app
os::cmd::expect_success_and_text 'oc new-app -f test/testdata/template-with-app-label.json -o yaml' 'app: myapp'

# verify that a template can be passed in stdin
os::cmd::expect_success 'cat examples/sample-app/application-template-stibuild.json | oc new-app -o yaml -f -'

# check search
os::cmd::expect_success_and_text 'oc new-app --search mysql' "Tags:\s+5.6, 5.7, latest"
os::cmd::expect_success_and_text 'oc new-app --search ruby-helloworld-sample' 'ruby-helloworld-sample'
# check search - partial matches
os::cmd::expect_success_and_text 'oc new-app --search ruby-hellow' 'ruby-helloworld-sample'
os::cmd::expect_success_and_text 'oc new-app --search --template=ruby-hel' 'ruby-helloworld-sample'
os::cmd::expect_success_and_text 'oc new-app --search --template=ruby-helloworld-sam -o yaml' 'ruby-helloworld-sample'
os::cmd::expect_success_and_text 'oc new-app --search rub' "Tags:\s+2.2, 2.3, latest"
os::cmd::expect_success_and_text 'oc new-app --search --image-stream=rub' "Tags:\s+2.2, 2.3, latest"
# check search - check correct usage of filters
os::cmd::expect_failure_and_not_text 'oc new-app --search --image-stream=ruby-heloworld-sample' 'application-template-stibuild'
os::cmd::expect_failure 'oc new-app --search --template=php'
os::cmd::expect_failure 'oc new-app -S --template=nodejs'
os::cmd::expect_failure 'oc new-app -S --template=perl'
# check search - filtered, exact matches
# make sure the imagestreams are imported first.
os::cmd::try_until_success 'oc get imagestreamtags mongodb:latest'
os::cmd::try_until_success 'oc get imagestreamtags mongodb:2.4'
os::cmd::try_until_success 'oc get imagestreamtags mongodb:2.6'
os::cmd::try_until_success 'oc get imagestreamtags mongodb:3.2'
os::cmd::try_until_success 'oc get imagestreamtags mysql:latest'
os::cmd::try_until_success 'oc get imagestreamtags mysql:5.5'
os::cmd::try_until_success 'oc get imagestreamtags mysql:5.6'
os::cmd::try_until_success 'oc get imagestreamtags mysql:5.7'
os::cmd::try_until_success 'oc get imagestreamtags nodejs:latest'
os::cmd::try_until_success 'oc get imagestreamtags nodejs:0.10'
os::cmd::try_until_success 'oc get imagestreamtags nodejs:4'
os::cmd::try_until_success 'oc get imagestreamtags perl:latest'
os::cmd::try_until_success 'oc get imagestreamtags perl:5.16'
os::cmd::try_until_success 'oc get imagestreamtags perl:5.20'
os::cmd::try_until_success 'oc get imagestreamtags perl:5.24'
os::cmd::try_until_success 'oc get imagestreamtags php:latest'
os::cmd::try_until_success 'oc get imagestreamtags php:5.5'
os::cmd::try_until_success 'oc get imagestreamtags php:5.6'
os::cmd::try_until_success 'oc get imagestreamtags postgresql:latest'
os::cmd::try_until_success 'oc get imagestreamtags postgresql:9.2'
os::cmd::try_until_success 'oc get imagestreamtags postgresql:9.4'
os::cmd::try_until_success 'oc get imagestreamtags postgresql:9.5'
os::cmd::try_until_success 'oc get imagestreamtags python:latest'
os::cmd::try_until_success 'oc get imagestreamtags python:2.7'
os::cmd::try_until_success 'oc get imagestreamtags python:3.3'
os::cmd::try_until_success 'oc get imagestreamtags python:3.4'
os::cmd::try_until_success 'oc get imagestreamtags python:3.5'
os::cmd::try_until_success 'oc get imagestreamtags ruby:latest'
os::cmd::try_until_success 'oc get imagestreamtags ruby:2.0'
os::cmd::try_until_success 'oc get imagestreamtags ruby:2.2'
os::cmd::try_until_success 'oc get imagestreamtags ruby:2.3'
os::cmd::try_until_success 'oc get imagestreamtags wildfly:latest'
os::cmd::try_until_success 'oc get imagestreamtags wildfly:10.1'
os::cmd::try_until_success 'oc get imagestreamtags wildfly:10.0'
os::cmd::try_until_success 'oc get imagestreamtags wildfly:9.0'
os::cmd::try_until_success 'oc get imagestreamtags wildfly:8.1'

os::cmd::expect_success_and_text 'oc new-app --search --image-stream=mongodb' "Tags:\s+2.6, 3.2, latest"
os::cmd::expect_success_and_text 'oc new-app --search --image-stream=mysql' "Tags:\s+5.6, 5.7, latest"
os::cmd::expect_success_and_text 'oc new-app --search --image-stream=nodejs' "Tags:\s+4, 6, latest"
os::cmd::expect_success_and_text 'oc new-app --search --image-stream=perl' "Tags:\s+5.20, 5.24, latest"
os::cmd::expect_success_and_text 'oc new-app --search --image-stream=php' "Tags:\s+5.6, 7.0, latest"
os::cmd::expect_success_and_text 'oc new-app --search --image-stream=postgresql' "Tags:\s+9.4, 9.5, latest"
os::cmd::expect_success_and_text 'oc new-app -S --image-stream=python' "Tags:\s+2.7, 3.4, 3.5, latest"
os::cmd::expect_success_and_text 'oc new-app -S --image-stream=ruby' "Tags:\s+2.2, 2.3, latest"
os::cmd::expect_success_and_text 'oc new-app -S --image-stream=wildfly' "Tags:\s+10.0, 10.1, 8.1, 9.0, latest"
os::cmd::expect_success_and_text 'oc new-app --search --template=ruby-helloworld-sample' 'ruby-helloworld-sample'
# check search - no matches
os::cmd::expect_failure_and_text 'oc new-app -S foo-the-bar' 'no matches found'
os::cmd::expect_failure_and_text 'oc new-app --search winter-is-coming' 'no matches found'
# check search - mutually exclusive flags
os::cmd::expect_failure_and_text 'oc new-app -S mysql --env=FOO=BAR' "can't be used"
os::cmd::expect_failure_and_text 'oc new-app --search mysql --code=https://github.com/openshift/ruby-hello-world' "can't be used"
os::cmd::expect_failure_and_text 'oc new-app --search mysql --param=FOO=BAR' "can't be used"

# set context-dir
os::cmd::expect_success_and_text 'oc new-app https://github.com/openshift/sti-ruby.git --context-dir="2.3/test/puma-test-app" -o yaml' 'contextDir: 2.3/test/puma-test-app'
os::cmd::expect_success_and_text 'oc new-app ruby~https://github.com/openshift/sti-ruby.git --context-dir="2.3/test/puma-test-app" -o yaml' 'contextDir: 2.3/test/puma-test-app'

# set strategy
os::cmd::expect_success_and_text 'oc new-app ruby~https://github.com/openshift/ruby-hello-world.git --strategy=docker -o yaml' 'dockerStrategy'
os::cmd::expect_success_and_text 'oc new-app https://github.com/openshift/ruby-hello-world.git --strategy=source -o yaml' 'sourceStrategy'

# prints volume and root user info
os::cmd::expect_success_and_text 'oc new-app --dry-run mysql' 'This image declares volumes'
os::cmd::expect_success_and_not_text 'oc new-app --dry-run mysql' "runs as the 'root' user"
os::cmd::expect_success_and_text 'oc new-app --dry-run --docker-image=mysql' 'This image declares volumes'
os::cmd::expect_success_and_text 'oc new-app --dry-run --docker-image=mysql' "WARNING: Image \"mysql\" runs as the 'root' user"

# verify multiple errors are displayed together, a nested error is returned, and that the usage message is displayed
os::cmd::expect_failure_and_text 'oc new-app --dry-run __template_fail __templatefile_fail' 'error: no match for "__template_fail"'
os::cmd::expect_failure_and_text 'oc new-app --dry-run __template_fail __templatefile_fail' 'error: no match for "__templatefile_fail"'
os::cmd::expect_failure_and_text 'oc new-app --dry-run __template_fail __templatefile_fail' 'error: unable to find the specified template file'
os::cmd::expect_failure_and_text 'oc new-app --dry-run __template_fail __templatefile_fail' "The 'oc new-app' command will match arguments"

# verify partial match error
os::cmd::expect_failure_and_text 'oc new-app --dry-run mysq' 'error: only a partial match was found for "mysq"'
os::cmd::expect_failure_and_text 'oc new-app --dry-run mysq' 'The argument "mysq" only partially matched'
os::cmd::expect_failure_and_text 'oc new-app --dry-run mysq' "Image stream \"mysql\" \\(tag \"5.7\"\\) in project"

# ensure new-app with pr ref does not fail
os::cmd::expect_success 'oc new-app https://github.com/openshift/ruby-hello-world#refs/pull/58/head --dry-run'

# verify image streams with no tags are reported correctly and that --allow-missing-imagestream-tags works
# new-app
os::cmd::expect_success 'printf "apiVersion: v1\nkind: ImageStream\nmetadata:\n  name: emptystream\n" | oc create -f -'
os::cmd::expect_failure_and_text 'oc new-app --dry-run emptystream' 'error: no tags found on matching image stream'
os::cmd::expect_success 'oc new-app --dry-run emptystream --allow-missing-imagestream-tags'
# new-build
os::cmd::expect_failure_and_text 'oc new-build --dry-run emptystream~https://github.com/openshift/ruby-ex' 'error: no tags found on matching image stream'
os::cmd::expect_success 'oc new-build --dry-run emptystream~https://github.com/openshift/ruby-ex --allow-missing-imagestream-tags --strategy=source'

# Allow setting --name when specifying grouping
os::cmd::expect_success "oc new-app mysql+ruby~https://github.com/openshift/ruby-ex --name foo -o yaml"
# but not with multiple components
os::cmd::expect_failure_and_text "oc new-app mysql ruby~https://github.com/openshift/ruby-ex --name foo -o yaml" "error: only one component or source repository can be used when specifying a name"
# do not allow specifying output image when specifying multiple input repos
os::cmd::expect_failure_and_text 'oc new-build https://github.com/openshift/nodejs-ex https://github.com/openshift/ruby-ex --to foo' 'error: only one component with source can be used when specifying an output image reference'
# but succeed with multiple input repos and no output image specified
os::cmd::expect_success 'oc new-build https://github.com/openshift/nodejs-ex https://github.com/openshift/ruby-ex -o yaml'
# check that binary build with a builder image results in a source type build
os::cmd::expect_success_and_text 'oc new-build --binary --image-stream=ruby -o yaml' 'type: Source'
# check that binary build with a specific strategy uses that strategy regardless of the image type
os::cmd::expect_success_and_text 'oc new-build --binary --image=ruby --strategy=docker -o yaml' 'type: Docker'

# When only a single imagestreamtag exists, and it does not match the implicit default
# latest tag, new-app should fail.
# when latest exists, we default to it and match it.
os::cmd::expect_success 'oc new-app --image-stream ruby https://github.com/openshift/rails-ex --dry-run'
# when latest does not exist, there are multiple partial matches (2.2, 2.3)
os::cmd::expect_success 'oc delete imagestreamtag ruby:latest'
os::cmd::expect_failure_and_text 'oc new-app --image-stream ruby https://github.com/openshift/rails-ex --dry-run' 'error: multiple images or templates matched \"ruby\":'
# when only 2.3 exists, there is a single partial match (2.3)
os::cmd::expect_success 'oc delete imagestreamtag ruby:2.2'
os::cmd::expect_failure_and_text 'oc new-app --image-stream ruby https://github.com/openshift/rails-ex --dry-run' 'error: only a partial match was found for \"ruby\":'
# when the tag is specified explicitly, the operation is successful
os::cmd::expect_success 'oc new-app --image-stream ruby:2.3 https://github.com/openshift/rails-ex --dry-run'

os::cmd::expect_success 'oc delete imagestreams --all'

# check that we can create from the template without errors
os::cmd::expect_success_and_text 'oc new-app ruby-helloworld-sample -l app=helloworld' 'service "frontend" created'
os::cmd::expect_success 'oc delete all -l app=helloworld'
os::cmd::expect_success 'oc delete secret dbsecret'
os::cmd::expect_success_and_text 'oc new-app ruby-helloworld-sample -l app=helloworld -o name' 'service/frontend'
os::cmd::expect_success 'oc delete all -l app=helloworld'
os::cmd::expect_success 'oc delete secret dbsecret'
os::cmd::expect_success 'oc delete template ruby-helloworld-sample'
# override component names
os::cmd::expect_success_and_text 'oc new-app mysql --name=db' 'db'
os::cmd::expect_success 'oc new-app https://github.com/openshift/ruby-hello-world -l app=ruby'
os::cmd::expect_success 'oc delete all -l app=ruby'
# check for error when template JSON file has errors
jsonfile="${OS_ROOT}/test/testdata/invalid.json"
os::cmd::expect_failure_and_text "oc new-app '${jsonfile}'" "error: unable to load template file \"${jsonfile}\": json: line 0: invalid character '}' after object key"

# check new-build
os::cmd::expect_failure_and_text 'oc new-build mysql -o yaml' 'you must specify at least one source repository URL'
os::cmd::expect_success_and_text 'oc new-build mysql --binary -o yaml --to mysql:bin' 'type: Binary'
os::cmd::expect_success_and_text 'oc new-build mysql https://github.com/openshift/ruby-hello-world --strategy=docker -o yaml' 'type: Docker'
os::cmd::expect_failure_and_text 'oc new-build mysql https://github.com/openshift/ruby-hello-world --binary' 'specifying binary builds and source repositories at the same time is not allowed'
# binary builds cannot be created unless a builder image is specified.
os::cmd::expect_failure_and_text 'oc new-build --name mybuild --binary --strategy=source -o yaml' 'you must provide a builder image when using the source strategy with a binary build'
os::cmd::expect_success_and_text 'oc new-build --name mybuild centos/ruby-22-centos7 --binary --strategy=source -o yaml' 'name: ruby-22-centos7:latest'
# binary builds can be created with no builder image if no strategy or docker strategy is specified
os::cmd::expect_success_and_text 'oc new-build --name mybuild --binary -o yaml' 'type: Binary'
os::cmd::expect_success_and_text 'oc new-build --name mybuild --binary --strategy=docker -o yaml' 'type: Binary'

# new-build image source tests
os::cmd::expect_failure_and_text 'oc new-build mysql --source-image centos' 'error: --source-image-path must be specified when --source-image is specified.'
os::cmd::expect_failure_and_text 'oc new-build mysql --source-image-path foo' 'error: --source-image must be specified when --source-image-path is specified.'

# ensure circular ref flagged but allowed for template
os::cmd::expect_success 'oc create -f test/testdata/circular-is.yaml'
os::cmd::expect_success_and_text 'oc new-app -f test/testdata/circular.yaml' 'should be different than input'
# ensure circular does not choke on image stream image
os::cmd::expect_success_and_not_text 'oc new-app -f test/testdata/bc-from-imagestreamimage.json --dry-run' 'Unable to follow reference type'

# do not allow use of non-existent image (should fail)
os::cmd::expect_failure_and_text 'oc new-app  openshift/bogusimage https://github.com/openshift/ruby-hello-world.git -o yaml' "no match for"
# allow use of non-existent image (should succeed)
os::cmd::expect_success 'oc new-app openshift/bogusimage https://github.com/openshift/ruby-hello-world.git -o yaml --allow-missing-images'

os::cmd::expect_success 'oc create -f test/testdata/installable-stream.yaml'

project=$(oc project -q)
os::cmd::expect_success 'oc policy add-role-to-user edit test-user'
os::cmd::expect_success 'oc login -u test-user -p anything'
os::cmd::try_until_success 'oc project ${project}'

os::cmd::try_until_success 'oc get imagestreamtags installable:file'
os::cmd::try_until_success 'oc get imagestreamtags installable:token'
os::cmd::try_until_success 'oc get imagestreamtags installable:serviceaccount'
os::cmd::expect_failure 'oc new-app installable:file'
os::cmd::expect_failure_and_text 'oc new-app installable:file' 'requires that you grant the image access'
os::cmd::expect_failure_and_text 'oc new-app installable:serviceaccount' "requires an 'installer' service account with project editor access"
os::cmd::expect_success_and_text 'oc new-app installable:file --grant-install-rights -o yaml' '/var/run/openshift.secret.token'
os::cmd::expect_success_and_text 'oc new-app installable:file --grant-install-rights -o yaml' 'activeDeadlineSeconds: 14400'
os::cmd::expect_success_and_text 'oc new-app installable:file --grant-install-rights -o yaml' 'openshift.io/generated-job: "true"'
os::cmd::expect_success_and_text 'oc new-app installable:file --grant-install-rights -o yaml' 'openshift.io/generated-job.for: installable:file'
os::cmd::expect_success_and_text 'oc new-app installable:token --grant-install-rights -o yaml' 'name: TOKEN_ENV'
os::cmd::expect_success_and_text 'oc new-app installable:token --grant-install-rights -o yaml' 'openshift/origin@sha256:'
os::cmd::expect_success_and_text 'oc new-app installable:serviceaccount --grant-install-rights -o yaml' 'serviceAccountName: installer'
os::cmd::expect_success_and_text 'oc new-app installable:serviceaccount --grant-install-rights -o yaml' 'fieldPath: metadata.namespace'
os::cmd::expect_success_and_text 'oc new-app installable:serviceaccount --grant-install-rights -o yaml A=B' 'name: A'

# Ensure output is valid JSON
os::cmd::expect_success 'oc new-app mongo -o json | python -m json.tool'

# Ensure custom branch/ref works
os::cmd::expect_success 'oc new-app https://github.com/openshift/ruby-hello-world#beta4'

# Ensure the resulting BuildConfig doesn't have unexpected sources
os::cmd::expect_success_and_not_text 'oc new-app https://github.com/openshift/ruby-hello-world --output-version=v1 -o=jsonpath="{.items[?(@.kind==\"BuildConfig\")].spec.source}"' 'dockerfile|binary'

# We permit running new-app against a remote URL which returns a template
os::cmd::expect_success 'oc new-app https://raw.githubusercontent.com/openshift/origin/master/examples/wordpress/template/wordpress-mysql.json --dry-run'

echo "new-app: ok"
os::test::junit::declare_suite_end
