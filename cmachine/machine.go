/*
 * Copyright 2019, Offchain Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmachine

/*
#cgo CFLAGS: -I.
#cgo LDFLAGS: -L. -lcavm -lavm -lstdc++
#include <cmachine.h>
#include <stdio.h>
#include <stdlib.h>
*/
import "C"

import (
	"bytes"
	"runtime"
	"unsafe"

	"github.com/offchainlabs/arb-util/machine"
	"github.com/offchainlabs/arb-util/protocol"
	"github.com/offchainlabs/arb-util/value"
)

type Machine struct {
	c unsafe.Pointer
}

func New(codeFile string) *Machine {
	cFilename := C.CString(codeFile)

	cMachine := C.machineCreate(cFilename)
	ret := &Machine{cMachine}
	runtime.SetFinalizer(ret, cdestroyVM)
	C.free(unsafe.Pointer(cFilename))
	return ret
}

func cdestroyVM(cMachine *Machine) {
	C.machineDestroy(cMachine.c)
}

func (m *Machine) Hash() (ret [32]byte) {
	C.machineHash(m.c, unsafe.Pointer(&ret[0]))
	return
}


func (m *Machine) Clone() machine.Machine {
	return &Machine{C.machineClone(m.c)}
}

func (m *Machine) InboxHash() (value.HashOnlyValue) {
	var hash [32]byte
	C.machineInboxHash(m.c, unsafe.Pointer(&hash[0]))
	return value.NewHashOnlyValue(hash, 0)
}

func (m *Machine) HasPendingMessages() bool {
	return C.machineHasPendingMessages(m.c) != 0
}

func (m *Machine) SendOnchainMessage(msg protocol.Message) {
	var buf bytes.Buffer
	err := value.MarshalValue(msg.AsValue(), &buf)
	if err != nil {
		panic(err)
	}
	msgData := buf.Bytes()
	C.machineSendOnchainMessage(m.c, unsafe.Pointer(&msgData[0]))
}

func (m *Machine) DeliverOnchainMessage() {
	C.machineDeliverOnchainMessages(m.c)
}

func (m *Machine) SendOffchainMessages(msgs[]protocol.Message) {
	var buf bytes.Buffer
	for _, msg := range msgs {
		err := value.MarshalValue(msg.AsValue(), &buf)
		if err != nil {
			panic(err)
		}
	}
	msgsData := buf.Bytes()
	if len(msgsData) > 0 {
		C.machineSendOffchainMessages(m.c, unsafe.Pointer(&msgsData[0]), C.int(len(msgs)))
	}
}

func (m *Machine) ExecuteAssertion(maxSteps int32, timeBounds protocol.TimeBounds) *protocol.Assertion {
	assertion := C.machineExecuteAssertion(
		m.c,
		C.uint64_t(maxSteps),
		C.uint64_t(timeBounds[0]),
		C.uint64_t(timeBounds[1]),
	)

	outMessagesRaw := C.GoBytes(unsafe.Pointer(assertion.outMessageData), assertion.outMessageLength)
	logsRaw := C.GoBytes(unsafe.Pointer(assertion.logData), assertion.logLength)

	outMessageVals := bytesArrayToVals(outMessagesRaw, int(assertion.outMessageCount))
	outMessages := make([]protocol.Message, 0, len(outMessageVals))
	for _, msgVal := range outMessageVals {
		msg, err := protocol.NewMessageFromValue(msgVal)
		if err != nil {
			panic(err)
		}
		outMessages = append(outMessages, msg)
	}

	logVals := bytesArrayToVals(logsRaw, int(assertion.logCount))


	return protocol.NewAssertion(
		m.Hash(),
		uint32(assertion.numSteps),
		outMessages,
		logVals,
	)
}

func (m *Machine) MarshalForProof() ([]byte, error) {
	rawProof := C.machineMarshallForProof(m.c)
	return C.GoBytes(unsafe.Pointer(rawProof.data), rawProof.length), nil
}

func bytesArrayToVals(data []byte, valCount int) []value.Value {
	rd := bytes.NewReader(data)
	vals := []value.Value{}
	for i := 0; i < valCount; i++ {
		val, err := value.UnmarshalValue(rd)
		if err != nil {
			panic(err)
		}
		vals = append(vals, val)
	}
	return vals
}