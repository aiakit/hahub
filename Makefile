.PHONY: proto data build

proto:
	for d in proto; do \
		for f in $$d/**/*.proto; do \
		    protoc  --ava_out=plugins=ava:. $$f; \
			echo compiled: $$f; \
		done; \
	done

d:
	export GOOS=linux; \
	./bin/build.sh; \
	./bin/deploy.sh; \
	./bin/run.sh

# build
build:
	./bin/build.sh


# build and stop ,then restart
run:
	./bin/build.sh; \
	./bin/stop.sh;\
	./bin/restart.sh;\

# stop
stop:
	./bin/stop.sh

# restart
restart:
	./bin/restart.sh

p:
	git add *; \
	git commit -m "fix";\
	git push;\
