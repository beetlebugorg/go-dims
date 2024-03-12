all:
	go build -o ./build/dims go-dims.go

static:
	go build -o ./build/dims -ldflags "-linkmode 'external' -extldflags '-fno-PIC -static -Wl,-z,stack-size=8388608 -lpng -lz -ltiff -lzstd -lwebp -lwebpmux -lwebpdemux -ljpeg -lbz2 -lfontconfig -lfreetype -lexpat -lbrotlidec -lbrotlienc -lbrotlicommon -llcms2 -lgomp'" go-dims.go

publish-lambda: docker
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin ${AWS_ACCOUNT_ID}.dkr.ecr.us-east-1.amazonaws.com
	docker tag beetlebugorg/go-dims:local ${AWS_ACCOUNT_ID}.dkr.ecr.us-east-1.amazonaws.com/beetlebugorg/go-dims:latest
	docker push ${AWS_ACCOUNT_ID}.dkr.ecr.us-east-1.amazonaws.com/beetlebugorg/go-dims:latest

create-function: publish-lambda
	aws lambda create-function --function-name go-dims --package-type Image --code ImageUri=${GO_DIMS_IMAGE} --role ${GO_DIMS_ROLE} --region ${GO_DIMS_REGION} --memory-size 1024 --environment "Variables={DIMS_SIGNING_KEY=${DIMS_SIGNING_KEY}}" --architectures arm64 --timeout 10


update-function: publish-lambda
	aws lambda update-function-code --function-name go-dims --image-uri ${GO_DIMS_IMAGE} --region ${GO_DIMS_REGION} --architectures arm64

docs:
	mdbook build docs

docs-serve:
	mdbook serve docs

docker: Dockerfile
	docker buildx build --load -t beetlebugorg/go-dims:local .
	docker images | grep beetlebugorg/go-dims