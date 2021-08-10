#!/usr/bin/env bash
# workaround for https://github.com/openshift/origin/issues/18942
OC="/lib/ld-musl-x86_64.so.1 --library-path /lib /usr/local/bin/oc"

if [[ $1 == "--config" ]] ; then
  cat <<EOF
configVersion: v1
kubernetes:
- apiVersion: user.openshift.io/v1
  kind: User
  executeHookOnEvent: ["Added"]
EOF
  exit 0
fi

# Adds a OC User to a Group
# Adds the label gepaplexx.com/groupsynced to the User, after it's synced
function addUserToGroup {
  userName=$1
  identity=$2

  echo "Adding ${userName} to group ${identity}"

  is_group=$(${OC} get Groups "${identity}" --ignore-not-found)
  if [[ ! $is_group ]]; then
    echo "Group ${identity} does not exist... creating group"
    ${OC} adm groups new "${identity}"
  fi
  ${OC} adm groups add-users "${identity}" "${userName}"
  ${OC} label User "${userName}" gepaplexx.com/groupsynced=true
  echo "Success"
}

function main {
  if [[ -z $BINDING_CONTEXT_PATH ]]; then
    echo "No binding found"
    exit 0
  fi

  type=$(jq -r '.[0].type' $BINDING_CONTEXT_PATH)

  # handle Synchronization event
  if [[ $type == "Synchronization" ]] ; then
    echo "Got Synchronization event"
    # select objects (Users) without label gepaplexx.com/groupsynced
    objects=$(jq -r '.[0].objects[].object | select(.metadata.labels["gepaplexx.com/groupsynced"]!="true")' $BINDING_CONTEXT_PATH | jq -s)
    ARRAY_COUNT=$(echo "${objects}" | jq -r '. | length-1')
    for IND in `seq 0 $ARRAY_COUNT`
    do
      userName=$(echo ${objects} | jq -r .[$IND].metadata.name)
      identity=$(echo ${objects} | jq -r .[$IND].identities[0])
      identity=($(echo ${identity} | tr ':' '\n'))
      identity="${identity[0]}"

      addUserToGroup "${userName}" "${identity}"
    done

  else # handle Hook event: User added
    echo "Got Hook event"
    ARRAY_COUNT=`jq -r '. | length-1' $BINDING_CONTEXT_PATH`
    for IND in `seq 0 $ARRAY_COUNT`
    do
      userName=$(jq -r .[$IND].object.metadata.name $BINDING_CONTEXT_PATH)
      identity=$(jq -r .[$IND].object.identities[0] $BINDING_CONTEXT_PATH)
      identity=($(echo ${identity} | tr ':' '\n'))
      identity="${identity[0]}"

      addUserToGroup "${userName}" "${identity}"
    done
  fi
}
main "$@"