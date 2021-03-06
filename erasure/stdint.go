/*
 * Mini Object Storage, (C) 2014 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package erasure

//
// int SizeInt()
// {
//      return sizeof(int);
// }
import "C"
import "unsafe"

var (
	// See http://golang.org/ref/spec#Numeric_types

	// SizeUint8 is the byte size of a uint8.
	SizeUint8 = int(unsafe.Sizeof(uint8(0)))
	// SizeUint16 is the byte size of a uint16.
	SizeUint16 = int(unsafe.Sizeof(uint16(0)))
	// SizeUint32 is the byte size of a uint32.
	SizeUint32 = int(unsafe.Sizeof(uint32(0)))
	// SizeUint64 is the byte size of a uint64.
	SizeUint64 = int(unsafe.Sizeof(uint64(0)))

	SizeInt = int(C.SizeInt())
	// SizeInt8  is the byte size of a int8.
	SizeInt8 = int(unsafe.Sizeof(int8(0)))
	// SizeInt16 is the byte size of a int16.
	SizeInt16 = int(unsafe.Sizeof(int16(0)))
	// SizeInt32 is the byte size of a int32.
	SizeInt32 = int(unsafe.Sizeof(int32(0)))
	// SizeInt64 is the byte size of a int64.
	SizeInt64 = int(unsafe.Sizeof(int64(0)))
)
