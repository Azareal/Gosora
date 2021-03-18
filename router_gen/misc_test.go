package main

import (
	"runtime/debug"
	"testing"
)

func exp(t *testing.T) func(bool, string) {
	return func(val bool, msg string) {
		if !val {
			debug.PrintStack()
			t.Error(msg)
		}
	}
}

func expf(t *testing.T) func(bool, string, ...interface{}) {
	return func(val bool, msg string, params ...interface{}) {
		if !val {
			debug.PrintStack()
			t.Errorf(msg, params...)
		}
	}
}

func TestPerc(t *testing.T) {
	ex, _, prec := exp(t), expf(t), NewPrec()
	ex(!prec.GreaterThan("MemberOnly", "AdminOnly"), "MemberOnly should not be greater then AdminOnly")
	ex(!prec.GreaterThan("AdminOnly", "MemberOnly"), "MemberOnly should not be greater then AdminOnly")
	ex(!prec.GreaterThan("NotInSet", "AdminOnly"), "NotInSet should not be greater then AdminOnly")
	ex(!prec.GreaterThan("AdminOnly", "NotInSet"), "AdminOnly should not be greater then NotInSet")
	ex(!prec.InAnySet("MemberOnly"), "MemberOnly should not be in any set")
	ex(!prec.InSameSet("MemberOnly", "AdminOnly"), "MemberOnly and AdminOnly should not be in the same set")
	ex(!prec.InSameSet("MemberOnly", "NotInSet"), "MemberOnly and NotInSet should not be in the same set")

	prec.AddSet("MemberOnly", "SuperModOnly", "AdminOnly", "SuperAdminOnly")
	ex(!prec.GreaterThan("MemberOnly", "AdminOnly"), "MemberOnly should not be greater then AdminOnly")
	ex(prec.GreaterThan("AdminOnly", "MemberOnly"), "AdminOnly should be greater then MemberOnly")
	ex(!prec.GreaterThan("NotInSet", "AdminOnly"), "NotInSet should not be greater then AdminOnly")
	ex(!prec.GreaterThan("AdminOnly", "NotInSet"), "AdminOnly should not be greater then NotInSet")
	ex(prec.InAnySet("MemberOnly"), "MemberOnly should be in a set")
	ex(!prec.InAnySet("NotInSet"), "NotInSet should not be in any set")
	ex(prec.InSameSet("MemberOnly", "AdminOnly"), "MemberOnly and AdminOnly should be in the same set")
	ex(!prec.InSameSet("MemberOnly", "NotInSet"), "MemberOnly and NotInSet should not be in the same set")

	items := prec.LessThanItem("AdminOnly")
	ex(len(items) > 0, "There should be items which are of a lower precedence than AdminOnly")
	imap := make(map[string]bool)
	for _, item := range items {
		imap[item] = true
	}
	mex := func(n string, val bool, msg string) {
		_, ok := imap[n]
		ex(ok == val, msg)
	}
	mex("SuperModOnly", true, "SuperModOnly should be returned in a list of lower precedence items than AdminOnly")
	mex("MemberOnly", true, "MemberOnly should be returned in a list of lower precedence items than AdminOnly")
	mex("SuperAdminOnly", false, "SuperAdminOnly should not be returned in a list of lower precedence items than AdminOnly")
	mex("NotInSet", false, "NotInSet should not be returned in a list of lower precedence items than AdminOnly")
}
