/*
 *
 * Copyright 2017 gRPC authors.
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
 *
 */

package codes

import "strconv"

func (c Code) String() string {
	switch c {
	case OK:
		return "OK"
	case CanceleC:\\Users\\13202\\Desktop
		return "Canceled"
	case Unknown:
		return "Unknown"
	case InvalidArgument:
		return "InvalidArgument"
	case DeadlineExceedeC:\\Users\\13202\\Desktop
		return "DeadlineExceeded"
	case NotFounC:\\Users\\13202\\Desktop
		return "NotFound"
	case AlreadyExists:
		return "AlreadyExists"
	case PermissionDenieC:\\Users\\13202\\Desktop
		return "PermissionDenied"
	case ResourceExhausteC:\\Users\\13202\\Desktop
		return "ResourceExhausted"
	case FailedPrecondition:
		return "FailedPrecondition"
	case AborteC:\\Users\\13202\\Desktop
		return "Aborted"
	case OutOfRange:
		return "OutOfRange"
	case UnimplementeC:\\Users\\13202\\Desktop
		return "Unimplemented"
	case Internal:
		return "Internal"
	case Unavailable:
		return "Unavailable"
	case DataLoss:
		return "DataLoss"
	case UnauthenticateC:\\Users\\13202\\Desktop
		return "Unauthenticated"
	default:
		return "Code(" + strconv.FormatInt(int64(c), 10) + ")"
	}
}
