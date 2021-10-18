set -e
cd  metacontroller
git clone https://github.com/metacontroller/metacontroller.git
cd  metacontroller
helm package deploy/helm/metacontroller --destination deploy/helm
helm upgrade --install metacontroller deploy/helm/metacontroller-v*.tgz
cd ..
rm -rf deploy
rm -rf metacontroller