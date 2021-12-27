segment=`cat template/segment`

count=$2

kubectl delete -f deploy
rm -rf deploy
mkdir deploy
cp template/deploy.yaml deploy/deploy.yaml
cp template/config.yaml deploy/config.yaml


for((i=0;i<$count;i++)); do
    export NUMBER=$i
    envsubst <./template/$1_segment> segment
    awk '{print "    "$0}' segment >> deploy/config.yaml
done

rm -f segment

kubectl apply  --server-side=true -f ./deploy/config.yaml
kubectl apply  -f ./deploy/deploy.yaml