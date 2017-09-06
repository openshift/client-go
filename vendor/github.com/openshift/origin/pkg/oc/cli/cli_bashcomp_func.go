package cli

const (
	bashCompletionFunc = `# call oc get $1,
__oc_override_flag_list=(config cluster user context namespace server)
__oc_override_flags()
{
    local ${__oc_override_flag_list[*]} two_word_of of
    for w in "${words[@]}"; do
        if [ -n "${two_word_of}" ]; then
            eval "${two_word_of}=\"--${two_word_of}=\${w}\""
            two_word_of=
            continue
        fi
        for of in "${__oc_override_flag_list[@]}"; do
            case "${w}" in
                --${of}=*)
                    eval "${of}=\"${w}\""
                    ;;
                --${of})
                    two_word_of="${of}"
                    ;;
            esac
        done
        if [ "${w}" == "--all-namespaces" ]; then
            namespace="--all-namespaces"
        fi
    done
    for of in "${__oc_override_flag_list[@]}"; do
        if eval "test -n \"\$${of}\""; then
            eval "echo \${${of}}"
        fi
    done
}
__oc_parse_get()
{

    local template
    template="{{ range .items  }}{{ .metadata.name }} {{ end }}"
    local oc_out
    if oc_out=$(oc get $(__oc_override_flags) -o template --template="${template}" "$1" 2>/dev/null); then
        COMPREPLY=( $( compgen -W "${oc_out[*]}" -- "$cur" ) )
    fi
}

__oc_get_namespaces()
{
    local template oc_out
    template="{{ range .items  }}{{ .metadata.name }} {{ end }}"
    if oc_out=$(oc get -o template --template="${template}" namespace 2>/dev/null); then
        COMPREPLY=( $( compgen -W "${oc_out[*]}" -- "$cur" ) )
    fi
}

__oc_get_resource()
{
    if [[ ${#nouns[@]} -eq 0 ]]; then
        return 1
    fi
    __oc_parse_get ${nouns[${#nouns[@]} -1]}
}

# $1 is the name of the pod we want to get the list of containers inside
__oc_get_containers()
{
    local template
    template="{{ range .spec.containers  }}{{ .name }} {{ end }}"
    __debug ${FUNCNAME} "nouns are ${nouns[@]}"

    local len="${#nouns[@]}"
    if [[ ${len} -ne 1 ]]; then
        return
    fi
    local last=${nouns[${len} -1]}
    local oc_out
    if oc_out=$(oc get -o template --template="${template}" pods "${last}" 2>/dev/null); then
        COMPREPLY=( $( compgen -W "${oc_out[*]}" -- "$cur" ) )
    fi
}

# Require both a pod and a container to be specified
__oc_require_pod_and_container()
{
    if [[ ${#nouns[@]} -eq 0 ]]; then
        __oc_parse_get pods
        return 0
    fi;
    __oc_get_containers
    return 0
}

__custom_func() {
    case ${last_command} in
 
        # first arg is the kind according to ValidArgs, second is resource name
        oc_get | oc_describe | oc_delete | oc_label | oc_stop | oc_expose | oc_export | oc_patch | oc_annotate | oc_env | oc_edit | oc_volume | oc_scale )
            __oc_get_resource
            return
            ;;

        # first arg is a pod name
        oc_rsh | oc_exec)
            if [[ ${#nouns[@]} -eq 0 ]]; then
                __oc_parse_get pods
            fi;
            return
            ;;
 
        # first arg is a pod name, second is a container name
        oc_logs | oc_attach)
            __oc_require_pod_and_container
            return
            ;;
 
        # args other than the first are filenames
        oc_secrets_new)
            # Complete args other than the first as filenames
            if [[ ${#nouns[@]} -gt 0 ]]; then
                _filedir
            fi;
            return
            ;;
 
        # first arg is a build config name
        oc_start-build | oc_cancel-build)
            if [[ ${#nouns[@]} -eq 0 ]]; then
                __oc_parse_get buildconfigs
            fi;
            return
            ;;
 
        # first arg is a deployment config
        oc_deploy)
            if [[ ${#nouns[@]} -eq 0 ]]; then
                __oc_parse_get deploymentconfigs
            fi;
            return
            ;;
 
        # first arg is a deployment config OR deployment
        oc_rollback)
            if [[ ${#nouns[@]} -eq 0 ]]; then
                __oc_parse_get deploymentconfigs,replicationcontrollers
            fi;
            return
            ;;

        # first arg is a project name
        oc_project)
            if [[ ${#nouns[@]} -eq 0 ]]; then
                __oc_parse_get projects
            fi;
            return
            ;;
 
        # first arg is an image stream
        oc_import-image)
            if [[ ${#nouns[@]} -eq 0 ]]; then
                __oc_parse_get imagestreams
            fi;
            return
            ;;
 
        *)
            ;;
    esac
}
`
)
