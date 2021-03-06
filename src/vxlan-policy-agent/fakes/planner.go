// This file was generated by counterfeiter
package fakes

import (
	"sync"
	"vxlan-policy-agent/enforcer"
	"vxlan-policy-agent/poller"
)

type Planner struct {
	GetRulesAndChainStub        func() (enforcer.RulesWithChain, error)
	getRulesAndChainMutex       sync.RWMutex
	getRulesAndChainArgsForCall []struct{}
	getRulesAndChainReturns     struct {
		result1 enforcer.RulesWithChain
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *Planner) GetRulesAndChain() (enforcer.RulesWithChain, error) {
	fake.getRulesAndChainMutex.Lock()
	fake.getRulesAndChainArgsForCall = append(fake.getRulesAndChainArgsForCall, struct{}{})
	fake.recordInvocation("GetRulesAndChain", []interface{}{})
	fake.getRulesAndChainMutex.Unlock()
	if fake.GetRulesAndChainStub != nil {
		return fake.GetRulesAndChainStub()
	} else {
		return fake.getRulesAndChainReturns.result1, fake.getRulesAndChainReturns.result2
	}
}

func (fake *Planner) GetRulesAndChainCallCount() int {
	fake.getRulesAndChainMutex.RLock()
	defer fake.getRulesAndChainMutex.RUnlock()
	return len(fake.getRulesAndChainArgsForCall)
}

func (fake *Planner) GetRulesAndChainReturns(result1 enforcer.RulesWithChain, result2 error) {
	fake.GetRulesAndChainStub = nil
	fake.getRulesAndChainReturns = struct {
		result1 enforcer.RulesWithChain
		result2 error
	}{result1, result2}
}

func (fake *Planner) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getRulesAndChainMutex.RLock()
	defer fake.getRulesAndChainMutex.RUnlock()
	return fake.invocations
}

func (fake *Planner) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ poller.Planner = new(Planner)
