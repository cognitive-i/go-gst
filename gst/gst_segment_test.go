package gst_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tinyzimmer/go-gst/gst"
)

func TestSampleReturnsValidBuffer(t *testing.T) {
	sut := gst.NewEmptySample()

	buf1 := gst.NewEmptyBuffer()
	sut.SetBuffer(buf1)

	// We cannot expect equivalance between Go Buffer wrapper
	// and the one returned, however .Instance should match
	buf2 := sut.GetBuffer()
	assert.NotNil(t, buf2)
	assert.Equal(t, buf1.Instance(), buf2.Instance())
}

func TestSampleReturnsNilBuffer(t *testing.T) {
	sut := gst.NewEmptySample()
	assert.Nil(t, sut.GetBuffer())
}

func TestSampleReturnsValidBufferList(t *testing.T) {
	sut := gst.NewEmptySample()

	bufList1 := gst.NewBufferList(nil)
	sut.SetBufferList(bufList1)

	// We cannot expect equivalance between Go Buffer wrapper
	// and the one returned, however .Instance should match
	bufList2 := sut.GetBufferList()
	assert.NotNil(t, bufList2)
	assert.Equal(t, bufList1.Instance(), bufList2.Instance())
}

func TestSampleReturnsNilBufferList(t *testing.T) {
	sut := gst.NewEmptySample()
	assert.Nil(t, sut.GetBufferList())
}

func TestMain(m *testing.M) {
	gst.Init(nil)
	defer gst.Deinit()
	os.Exit(m.Run())
}
