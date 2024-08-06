package vp8decoder

import (
	"sync"
)

type FrameMeta struct {
	Timestamp uint32
}

type FrameContainer struct {
	// Decided to not use a byte slice here because it cannot be easily passed via URL parameters (unlike its string representation)
	FramesMap      map[string][][]byte
	MetadataMap    map[string][]FrameMeta
	PipelineMap    map[string]string
	DatachannelMap map[string]chan []byte
	lock           sync.Mutex
}

func NewFrameContainer() *FrameContainer {
	return &FrameContainer{
		FramesMap:      make(map[string][][]byte),
		MetadataMap:    make(map[string][]FrameMeta),
		DatachannelMap: make(map[string]chan []byte),
		PipelineMap:    make(map[string]string),
	}
}

func (fc *FrameContainer) AddFrame(resource, pipeline string, frame []byte, meta FrameMeta) {
	fc.lock.Lock()
	defer fc.lock.Unlock()
	fc.FramesMap[resource] = append([][]byte{frame}, fc.FramesMap[resource]...)
	fc.MetadataMap[resource] = append([]FrameMeta{meta}, fc.MetadataMap[resource]...)
	fc.PipelineMap[resource] = pipeline

	for len(fc.FramesMap[resource]) > 10 {
		// remove last:
		fc.FramesMap[resource] = fc.FramesMap[resource][:len(fc.FramesMap[resource])-1]
		fc.MetadataMap[resource] = fc.MetadataMap[resource][:len(fc.MetadataMap[resource])-1]
	}
}

func (fc *FrameContainer) GetFrames(resource string) ([][]byte, []FrameMeta) {
	fc.lock.Lock()
	defer fc.lock.Unlock()
	return fc.FramesMap[resource], fc.MetadataMap[resource]
}

// deduplicate (generics?)
// XXX might be stupid
func IsClosedDC(ch <-chan []byte) bool {
	select {
	case <-ch:
		return true
	default:
	}

	return false
}

func (fc *FrameContainer) RemoveResource(resource string) {
	if !fc.HasResource(resource) {
		return
	}
	fc.lock.Lock()
	defer fc.lock.Unlock()
	if fc.DatachannelMap[resource] != nil && !IsClosedDC(fc.DatachannelMap[resource]) {
		close(fc.DatachannelMap[resource])
	}
	delete(fc.DatachannelMap, resource)
	delete(fc.FramesMap, resource)
	delete(fc.MetadataMap, resource)
	delete(fc.PipelineMap, resource)
}

func (fc *FrameContainer) HasResource(resource string) bool {
	if fc.FramesMap == nil {
		return false
	}
	fc.lock.Lock()
	defer fc.lock.Unlock()
	_, ok := fc.FramesMap[resource]
	return ok
}

func (fc *FrameContainer) Resources() ([]string, []string) {
	fc.lock.Lock()
	defer fc.lock.Unlock()
	resources := make([]string, 0, len(fc.FramesMap))
	pipelines := make([]string, 0, len(fc.FramesMap))
	for k := range fc.FramesMap {
		resources = append(resources, k)
		pipelines = append(pipelines, fc.PipelineMap[k])
	}
	return resources, pipelines
}

func (fc *FrameContainer) GetDataChannel(resource string) chan []byte {
	fc.lock.Lock()
	defer fc.lock.Unlock()
	if fc.DatachannelMap[resource] == nil {
		fc.DatachannelMap[resource] = make(chan []byte, 10)
	}
	return fc.DatachannelMap[resource]
}
