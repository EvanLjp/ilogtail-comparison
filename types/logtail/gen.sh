
function makeConfig () {
    echo "{\"metrics\" : {" > temp.json
    count=$2
    for((i=0;i<$count;i++)); do
        export NUMBER=$i
        envsubst <./static/template/$1_segment > segment
        cat segment >> temp.json
    if (( $i != $count-1 )); then
        echo "," >> temp.json
    fi
    done
    echo "}}" >> temp.json

    jq -c . temp.json > user_local_config.json
    rm -rf temp.json
}


function transferToRemote () {
    pods=`kubectl get pods -n kube-system |grep logtail-ds|cut -d ' ' -f 1`
    for item in ${pods[@]}; do
        echo $item
        kubectl cp ./user_local_config.json kube-system/$item:/etc/ilogtail/user_local_config.json
    done

}


makeConfig $1 $2
transferToRemote
