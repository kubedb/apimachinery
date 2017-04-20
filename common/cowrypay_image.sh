#!/bin/bash

# override this one if you need to change push & pull
docker_push() {
	hub_canary cowrypay
}

docker_pull() {
	hub_pull cowrypay
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
			build cowrypay
			;;
		build_binary)
			build_binary
			;;
		build_docker)
			build_docker
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
		release)
			docker_release cowrypay
			;;
		check)
			docker_check cowrypay
			;;
		run)
			docker_run cowrypay
			;;
		sh)
			docker_sh cowrypay
			;;
		rm)
			docker_rm
			;;
		rmi)
			docker_rmi
			;;
		*)	(10)
			echo $"Usage: $0 {build|build_binary|build_docker|clean|push|pull|release|check|sh|rm|rmi}"
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
			build cowrypay
			;;
		clean)
			clean
			;;
		push)
			docker_up cowrypay $IMG:$TAG
			;;
		pull)
			docker_pull cowrypay
			;;
		release)
			docker_release cowrypay
			;;
		check)
			docker_check cowrypay
			;;
		run)
			docker_run cowrypay
			;;
		sh)
			docker_sh cowrypay
			;;
		rm)
			docker_rm
			;;
		rmi)
			docker_rmi
			;;
		*)	(10)
			echo $"Usage: $0 {build|clean|push|pull|release|check|sh|rm|rmi}"
			RETVAL=1
	esac
	exit $RETVAL
}
