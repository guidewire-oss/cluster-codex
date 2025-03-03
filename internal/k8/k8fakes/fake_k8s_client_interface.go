package k8fakes

import (
	"cluster-codex/internal/k8"
	"cluster-codex/internal/model"
	"context"
	"sync"
)

type FakeK8sClientInterface struct {
	GetAllComponentsStub        func(context.Context) ([]model.Component, error)
	getAllComponentsMutex       sync.RWMutex
	getAllComponentsArgsForCall []struct {
		arg1 context.Context
	}
	getAllComponentsReturns struct {
		result1 []model.Component
		result2 error
	}
	getAllComponentsReturnsOnCall map[int]struct {
		result1 []model.Component
		result2 error
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

// Implement GetAllComponents
func (fake *FakeK8sClientInterface) GetAllComponents(arg1 context.Context) ([]model.Component, error) {
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
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

// Implement GetAllImages
func (fake *FakeK8sClientInterface) GetAllImages(arg1 context.Context, arg2 []string) ([]model.Component, error) {
	fake.getAllImagesMutex.Lock()
	ret, specificReturn := fake.getAllImagesReturnsOnCall[len(fake.getAllImagesArgsForCall)]
	fake.getAllImagesArgsForCall = append(fake.getAllImagesArgsForCall, struct {
		arg1 context.Context
		arg2 []string
	}{arg1, arg2})
	stub := fake.GetAllImagesStub
	fakeReturns := fake.getAllImagesReturns
	fake.recordInvocation("GetAllImages", []interface{}{arg1, arg2})
	fake.getAllImagesMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

// GetAllImagesCallCount - Returns number of calls
func (fake *FakeK8sClientInterface) GetAllImagesCallCount() int {
	fake.getAllImagesMutex.RLock()
	defer fake.getAllImagesMutex.RUnlock()
	return len(fake.getAllImagesArgsForCall)
}

// GetAllImagesCalls - Sets a stub function
func (fake *FakeK8sClientInterface) GetAllImagesCalls(stub func(context.Context, []string) ([]model.Component, error)) {
	fake.getAllImagesMutex.Lock()
	defer fake.getAllImagesMutex.Unlock()
	fake.GetAllImagesStub = stub
}

// GetAllImagesArgsForCall - Returns the arguments for a call
func (fake *FakeK8sClientInterface) GetAllImagesArgsForCall(i int) (context.Context, []string) {
	fake.getAllImagesMutex.RLock()
	defer fake.getAllImagesMutex.RUnlock()
	argsForCall := fake.getAllImagesArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

// GetAllImagesReturns - Sets return values for all calls
func (fake *FakeK8sClientInterface) GetAllImagesReturns(result1 []model.Component, result2 error) {
	fake.getAllImagesMutex.Lock()
	defer fake.getAllImagesMutex.Unlock()
	fake.GetAllImagesStub = nil
	fake.getAllImagesReturns = struct {
		result1 []model.Component
		result2 error
	}{result1, result2}
}

// GetAllImagesReturnsOnCall - Sets return values for a specific call index
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

// Invocations - Records function calls
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

// Record function invocation
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

// Ensure FakeK8sClientInterface implements K8sClientInterface
var _ k8.K8sClientInterface = new(FakeK8sClientInterface)
