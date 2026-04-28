package raft

//import (
//	"errors"
//	"net"
//	"time"
//
//	"k8s.io/klog/v2"
//)
//
//// Accept waits for and returns the next connection with TCP keep-alive settings.
//// It can be interrupted by the stop channel for graceful shutdown handling.
//func (ln stoppableListener) Accept() (c net.Conn, err error) {
//	connc := make(chan *net.TCPConn, 1)
//	errc := make(chan error, 1)
//	go func() {
//		tc, err := ln.AcceptTCP()
//		if err != nil {
//			errc <- err
//			return
//		}
//		connc <- tc
//	}()
//	select {
//	case <-ln.stopc:
//		return nil, errors.New("server stopped")
//	case err := <-errc:
//		return nil, err
//	case tc := <-connc:
//		err = tc.SetKeepAlive(true)
//		if err != nil {
//			klog.Errorln(err)
//		}
//		err = tc.SetKeepAlivePeriod(3 * time.Minute)
//		if err != nil {
//			klog.Errorln(err)
//		}
//		return tc, nil
//	}
//}
