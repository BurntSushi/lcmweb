install:
	go install ./...

gofmt:
	gofmt -w *.go
	colcheck *.go

loc:
	find ./ -name '*.go' -print | sort | xargs wc -l

tags:
	find ./ -name '*.go' -print0 | xargs -0 gotags > TAGS

push:
	git push origin master
	git push github master

