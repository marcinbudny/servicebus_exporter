DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
pushd $DIR/charts
helm package servicebusexporter --destination=.deploy
cr upload -o giggio -r servicebus_exporter -p .deploy
popd
pushd $DIR
cr index -i index.yaml -p ./charts/.deploy/ -o giggio -r servicebus_exporter -c https://github.com/giggio/servicebus_exporter/
popd
