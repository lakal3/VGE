package vanimation

import "github.com/lakal3/vge/vge/vmodel"

type MapJoint func(sk *vmodel.Skin, name string) (jIndex int)

func DefaultMapJointfunc(sk *vmodel.Skin, name string) (jIndex int) {
	for idx, j := range sk.Joints {
		if j.Name == name {
			return idx
		}
	}
	return -1
}
