package bw2bind

import (
	"strings"
	"sync"
)

type ServiceClient Service
type InterfaceClient Interface

func (cl *BW2Client) NewServiceClient(baseuri string, name string) *ServiceClient {
	baseuri = strings.TrimSuffix(baseuri, "/")
	return &ServiceClient{cl: cl, baseuri: baseuri, name: name, mu: &sync.Mutex{}}
}

func (sc *ServiceClient) AddInterface(prefix string, name string) *InterfaceClient {
	svc := Service(*sc)
	ifc := Interface{
		svc:    &svc,
		prefix: prefix,
		name:   name,
		auto:   false,
	}
	sc.mu.Lock()
	sc.ifaces = append(sc.ifaces, &ifc)
	sc.mu.Unlock()

	rv := InterfaceClient(ifc)
	return &rv
}

func (sc *ServiceClient) GetMetadata() (map[string]*MetadataTuple, error) {
	svc := Service(*sc)
	md, _, err := sc.cl.GetMetadata(svc.FullURI())
	return md, err
}

func (sc *ServiceClient) GetMetadataKey(key string) (*MetadataTuple, error) {
	svc := Service(*sc)
	md, _, err := sc.cl.GetMetadataKey(svc.FullURI(), key)
	return md, err
}

func (ifclient *InterfaceClient) SignalURI(signal string) string {
	ifc := Interface(*ifclient)
	return ifc.SignalURI(signal)
}

func (ifclient *InterfaceClient) SlotURI(slot string) string {
	ifc := Interface(*ifclient)
	return ifc.SlotURI(slot)
}

func (ifclient *InterfaceClient) FullURI() string {
	ifc := Interface(*ifclient)
	return ifc.FullURI()
}

func (ifclient *InterfaceClient) PublishSlot(slot string, poz ...PayloadObject) error {
	return ifclient.svc.cl.Publish(&PublishParams{
		URI:            ifclient.SlotURI(slot),
		AutoChain:      true,
		PayloadObjects: poz,
	})
}

func (ifclient *InterfaceClient) SubscribeSignal(signal string, cb func(*SimpleMessage)) error {
	subChan, err := ifclient.svc.cl.Subscribe(&SubscribeParams{
		URI:       ifclient.SignalURI(signal),
		AutoChain: true,
	})
	if err != nil {
		return err
	}

	go func() {
		for sm := range subChan {
			cb(sm)
		}
	}()
	return nil
}

func (ifclient *InterfaceClient) SubscribeSignalH(signal string, cb func(*SimpleMessage)) (string, error) {
	subChan, handle, err := ifclient.svc.cl.SubscribeH(&SubscribeParams{
		URI:       ifclient.SignalURI(signal),
		AutoChain: true,
	})
	if err != nil {
		return "", err
	}

	go func() {
		for sm := range subChan {
			cb(sm)
		}
	}()
	return handle, nil
}
