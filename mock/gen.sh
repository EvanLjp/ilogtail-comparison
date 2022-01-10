kubectl delete -f deploy
rm -rf deploy
mkdir deploy

for((i=0;i<$1;i++)); do
    export NAME=nginx-log-demo-$i
    envsubst <mock.yaml>  deploy/$NAME.yaml
done


kubectl apply -f deploy