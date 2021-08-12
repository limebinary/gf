// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package gproc

import (
	"fmt"
	"github.com/gogf/gf/container/gmap"
	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/net/gtcp"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/util/gconv"
)

// MsgRequest is the request structure for process communication.
type MsgRequest struct {
	SendPid int    // Sender PID.
	RecvPid int    // Receiver PID.
	Group   string // Message group name.
	Data    []byte // Request data.
}

// MsgResponse is the response structure for process communication.
type MsgResponse struct {
	Code    int    // 1: OK; Other: Error.
	Message string // Response message.
	Data    []byte // Response data.
}

const (
	defaultGroupNameFoProcComm = ""    // Default group name.
	defaultTcpPortForProcComm  = 10000 // Starting port number for receiver listening.
	maxLengthForProcMsgQueue   = 10000 // Max size for each message queue of the group.
)

var (
	// commReceiveQueues is the group name to queue map for storing received data.
	// The value of the map is type of *gqueue.Queue.
	commReceiveQueues = gmap.NewStrAnyMap(true)

	// commPidFolderPath specifies the folder path storing pid to port mapping files.
	commPidFolderPath = gfile.TempDir("gproc")
)

func init() {
	// Automatically create the storage folder.
	if !gfile.Exists(commPidFolderPath) {
		err := gfile.Mkdir(commPidFolderPath)
		if err != nil {
			panic(fmt.Errorf(`create gproc folder failed: %v`, err))
		}
	}
}

// getConnByPid creates and returns a TCP connection for specified pid.
func getConnByPid(pid int) (*gtcp.PoolConn, error) {
	port := getPortByPid(pid)
	if port > 0 {
		if conn, err := gtcp.NewPoolConn(fmt.Sprintf("127.0.0.1:%d", port)); err == nil {
			return conn, nil
		} else {
			return nil, err
		}
	}
	return nil, gerror.NewCodef(gerror.CodeOperationFailed, "could not find port for pid: %d", pid)
}

// getPortByPid returns the listening port for specified pid.
// It returns 0 if no port found for the specified pid.
func getPortByPid(pid int) int {
	path := getCommFilePath(pid)
	content := gfile.GetContentsWithCache(path)
	return gconv.Int(content)
}

// getCommFilePath returns the pid to port mapping file path for given pid.
func getCommFilePath(pid int) string {
	return gfile.Join(commPidFolderPath, gconv.String(pid))
}
