#!/bin/bash

# override this one if you need to change push & pull
docker_push() {
	attic_up appscode
}

docker_pull() {
	attic_pull appscode
}


source_repo() {
	RETVAL=0

	if [ $# -eq 0 ]; then
		cmd=${DEFAULT_COMMAND:-build}
		$cmd
		exit $RETVAL
	fi

	case "$1" in
		build)
			build appscode
			;;
		build_binary)
			build_binary
			;;
		build_docker)
			build_docker
			;;
		build_docker_phd)
			build_docker_phd
			;;
		clean)
			clean
			;;
		push)
			docker_push
			;;
		pull)
			docker_pull
			;;
		gcr)
			gcr_pull appscode
			;;
		check)
			docker_check appscode
			;;
		run)
			docker_run appscode
			;;
		sh)
			docker_sh appscode
			;;
		rm)
			docker_rm
			;;
		rmi)
			docker_rmi
			;;
		*)	(10)
			echo $"Usage: $0 {build|build_binary|build_docker|clean|push|pull|check|sh|rm|rmi}"
			RETVAL=1
	esac
	exit $RETVAL
}

binary_repo() {
	RETVAL=0

	if [ $# -eq 0 ]; then
		cmd=${DEFAULT_COMMAND:-build}
		$cmd
		exit $RETVAL
	fi

	case "$1" in
		build)
			build appscode
			;;
		clean)
			clean
			;;
		push)
			docker_push
			;;
		pull)
			docker_pull
			;;
		gcr)
			gcr_pull appscode
			;;
		check)
			docker_check appscode
			;;
		run)
			docker_run appscode
			;;
		sh)
			docker_sh appscode
			;;
		rm)
			docker_rm
			;;
		rmi)
			docker_rmi
			;;
		*)	(10)
			echo $"Usage: $0 {build|clean|push|pull|check|sh|rm|rmi}"
			RETVAL=1
	esac
	exit $RETVAL
}
