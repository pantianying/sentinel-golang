package dubbo

import (
	"context"
	"github.com/apache/dubbo-go/protocol"
)
import (
	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
)

type providerFilter struct{}

func (d *providerFilter) Invoke(ctx context.Context, invoker protocol.Invoker, invocation protocol.Invocation) protocol.Result {
	methodResourceName := getResourceName(invoker, invocation, getProviderPrefix())
	interfaceResourceName := ""
	if getInterfaceGroupAndVersionEnabled() {
		interfaceResourceName = getColonSeparatedKey(invoker.GetUrl())
	} else {
		interfaceResourceName = invoker.GetUrl().Service()
	}
	_, b := sentinel.Entry(interfaceResourceName, sentinel.WithResourceType(base.ResTypeRPC), sentinel.WithTrafficType(base.Inbound))
	if b != nil { // blocked
		return providerDubboFallback(ctx, invoker, invocation, b)
	}
	_, b = sentinel.Entry(methodResourceName, sentinel.WithResourceType(base.ResTypeRPC), sentinel.WithTrafficType(base.Inbound), sentinel.WithArgs(invocation.Attachments()))
	if b != nil { // blocked
		return providerDubboFallback(ctx, invoker, invocation, b)
	}

	return invoker.Invoke(ctx, invocation)
}

func (d *providerFilter) OnResponse(ctx context.Context, result protocol.Result, _ protocol.Invoker, _ protocol.Invocation) protocol.Result {
	if interfaceEntry := ctx.Value(InterfaceEntryKey); interfaceEntry != nil {
		// TODO traceEntry()
		interfaceEntry.(*base.SentinelEntry).Exit()
	}
	if methodEntry := ctx.Value(MethodEntryKey); methodEntry != nil {
		// TODO traceEntry()
		methodEntry.(*base.SentinelEntry).Exit()
	}
	return result
}
