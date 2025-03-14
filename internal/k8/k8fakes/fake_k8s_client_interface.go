// Code generated by counterfeiter. DO NOT EDIT.
package k8fakes

import (
	"cluster-codex/internal/k8"
	"cluster-codex/internal/model"
	"context"
	"sync"
)

type FakeK8sClientInterface struct {
	GetAllComponentsStub        func(context.Context) ([]model.Component, []string, error)
	getAllComponentsMutex       sync.RWMutex
	getAllComponentsArgsForCall []struct {
		arg1 context.Context
	}
	getAllComponentsReturns struct {
		result1 []model.Component
		result2 []string
		result3 error
	}
	getAllComponentsReturnsOnCall map[int]struct {
		result1 []model.Component
		result2 []string
		result3 error
	}
	GetAllImagesStub        func(context.Context, []string) ([]model.Component, error)
	getAllImagesMutex       sync.RWMutex
	getAllImagesArgsForCall []struct {
		arg1 context.Context
		arg2 []string
	}
	getAllImagesReturns struct {
		result1 []model.Component
		result2 error
	}
	getAllImagesReturnsOnCall map[int]struct {
		result1 []model.Component
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeK8sClientInterface) GetAllComponents(arg1 context.Context) ([]model.Component, []string, error) {
	fake.getAllComponentsMutex.Lock()
	ret, specificReturn := fake.getAllComponentsReturnsOnCall[len(fake.getAllComponentsArgsForCall)]
	fake.getAllComponentsArgsForCall = append(fake.getAllComponentsArgsForCall, struct {
		arg1 context.Context
	}{arg1})
	stub := fake.GetAllComponentsStub
	fakeReturns := fake.getAllComponentsReturns
	fake.recordInvocation("GetAllComponents", []interface{}{arg1})
	fake.getAllComponentsMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2, ret.result3
	}
	return fakeReturns.result1, fakeReturns.result2, fakeReturns.result3
}

func (fake *FakeK8sClientInterface) GetAllComponentsCallCount() int {
	fake.getAllComponentsMutex.RLock()
	defer fake.getAllComponentsMutex.RUnlock()
	return len(fake.getAllComponentsArgsForCall)
}

func (fake *FakeK8sClientInterface) GetAllComponentsCalls(stub func(context.Context) ([]model.Component, []string, error)) {
	fake.getAllComponentsMutex.Lock()
	defer fake.getAllComponentsMutex.Unlock()
	fake.GetAllComponentsStub = stub
}

func (fake *FakeK8sClientInterface) GetAllComponentsArgsForCall(i int) context.Context {
	fake.getAllComponentsMutex.RLock()
	defer fake.getAllComponentsMutex.RUnlock()
	argsForCall := fake.getAllComponentsArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeK8sClientInterface) GetAllComponentsReturns(result1 []model.Component, result2 []string, result3 error) {
	fake.getAllComponentsMutex.Lock()
	defer fake.getAllComponentsMutex.Unlock()
	fake.GetAllComponentsStub = nil
	fake.getAllComponentsReturns = struct {
		result1 []model.Component
		result2 []string
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeK8sClientInterface) GetAllComponentsReturnsOnCall(i int, result1 []model.Component, result2 []string, result3 error) {
	fake.getAllComponentsMutex.Lock()
	defer fake.getAllComponentsMutex.Unlock()
	fake.GetAllComponentsStub = nil
	if fake.getAllComponentsReturnsOnCall == nil {
		fake.getAllComponentsReturnsOnCall = make(map[int]struct {
			result1 []model.Component
			result2 []string
			result3 error
		})
	}
	fake.getAllComponentsReturnsOnCall[i] = struct {
		result1 []model.Component
		result2 []string
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeK8sClientInterface) GetAllImages(arg1 context.Context, arg2 []string) ([]model.Component, error) {
	var arg2Copy []string
	if arg2 != nil {
		arg2Copy = make([]string, len(arg2))
		copy(arg2Copy, arg2)
	}
	fake.getAllImagesMutex.Lock()
	ret, specificReturn := fake.getAllImagesReturnsOnCall[len(fake.getAllImagesArgsForCall)]
	fake.getAllImagesArgsForCall = append(fake.getAllImagesArgsForCall, struct {
		arg1 context.Context
		arg2 []string
	}{arg1, arg2Copy})
	stub := fake.GetAllImagesStub
	fakeReturns := fake.getAllImagesReturns
	fake.recordInvocation("GetAllImages", []interface{}{arg1, arg2Copy})
	fake.getAllImagesMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeK8sClientInterface) GetAllImagesCallCount() int {
	fake.getAllImagesMutex.RLock()
	defer fake.getAllImagesMutex.RUnlock()
	return len(fake.getAllImagesArgsForCall)
}

func (fake *FakeK8sClientInterface) GetAllImagesCalls(stub func(context.Context, []string) ([]model.Component, error)) {
	fake.getAllImagesMutex.Lock()
	defer fake.getAllImagesMutex.Unlock()
	fake.GetAllImagesStub = stub
}

func (fake *FakeK8sClientInterface) GetAllImagesArgsForCall(i int) (context.Context, []string) {
	fake.getAllImagesMutex.RLock()
	defer fake.getAllImagesMutex.RUnlock()
	argsForCall := fake.getAllImagesArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeK8sClientInterface) GetAllImagesReturns(result1 []model.Component, result2 error) {
	fake.getAllImagesMutex.Lock()
	defer fake.getAllImagesMutex.Unlock()
	fake.GetAllImagesStub = nil
	fake.getAllImagesReturns = struct {
		result1 []model.Component
		result2 error
	}{result1, result2}
}

func (fake *FakeK8sClientInterface) GetAllImagesReturnsOnCall(i int, result1 []model.Component, result2 error) {
	fake.getAllImagesMutex.Lock()
	defer fake.getAllImagesMutex.Unlock()
	fake.GetAllImagesStub = nil
	if fake.getAllImagesReturnsOnCall == nil {
		fake.getAllImagesReturnsOnCall = make(map[int]struct {
			result1 []model.Component
			result2 error
		})
	}
	fake.getAllImagesReturnsOnCall[i] = struct {
		result1 []model.Component
		result2 error
	}{result1, result2}
}

func (fake *FakeK8sClientInterface) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getAllComponentsMutex.RLock()
	defer fake.getAllComponentsMutex.RUnlock()
	fake.getAllImagesMutex.RLock()
	defer fake.getAllImagesMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeK8sClientInterface) recordInvocation(key string, args []interface{}) {
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

var _ k8.K8sClientInterface = new(FakeK8sClientInterface)
