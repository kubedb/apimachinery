#!/bin/sh

# install curl
apk --update add curl

# required environment variables

TARGET_URL=http://127.0.0.1:9090/namespaces/${NAMESPACE}/workplans/${WORKPLAN}
REPO=kubedb/apimachinery
CONTEXT=kubeci

STATE_PENDING=pending
STATE_SUCCESS=success
STATE_FAILURE=failure
DESCRIPTION_PENDING=kubeci-build-pending
DESCRIPTION_SUCCESS=kubeci-build-success
DESCRIPTION_FAILURE=kubeci-build-failure

# HEAD_SHA
# TOKEN
# PR_NUMBER

# githhub status api post data

generate_status() {
	cat <<EOF
{
  "state": "$STATE",
  "target_url": "$TARGET_URL",
  "description": "$DESCRIPTION",
  "context": "$CONTEXT"
}
EOF
}

API_URL=https://api.github.com/repos/${REPO}/statuses/${HEAD_SHA}?access_token=${TOKEN}

# set pending status
echo "setting pending status..."

STATE=$STATE_PENDING
DESCRIPTION=$DESCRIPTION_PENDING
curl -s -o /dev/null -X POST -d "$(generate_status)" $API_URL

# link repository files to package path

mkdir -p $GOPATH/src/github.com/kubedb/
ln -s /kubeci/workspace $GOPATH/src/github.com/kubedb/apimachinery

# run tests
echo "running go tests..."

cd $GOPATH/src/github.com/kubedb/apimachinery
go test ./...

# set succeed/failed status

if [ $? -eq 0 ]; then
	STATE=$STATE_SUCCESS
	DESCRIPTION=$DESCRIPTION_SUCCESS
else
	STATE=$STATE_FAILURE
	DESCRIPTION=$DESCRIPTION_FAILURE
fi

echo "setting succeed/failed status..."
curl -s -o /dev/null -X POST -d "$(generate_status)" $API_URL

# remove ok-to-test label
echo "removing ok-to-test label..."

API_URL=https://api.github.com/repos/${REPO}/issues/${PR_NUMBER}/labels/ok-to-test?access_token=${TOKEN}
curl -s -o /dev/null -X DELETE $API_URL

echo "done"
