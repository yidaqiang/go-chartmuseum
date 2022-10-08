package chartmuseum

import (
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestInfoService_Index(t *testing.T) {
	convey.Convey("获取首页", t, func() {
		index, err := testClient.Info.Index()
		if err != nil {
			t.Error(err)
		}
		convey.So(index, convey.ShouldNotBeEmpty)
	})
}

func TestInfoService_Health(t *testing.T) {
	convey.Convey("获取健康状态", t, func() {
		health, err := testClient.Info.Health()
		if err != nil {
			t.Error(err)
		}
		convey.So(health.Healthy, convey.ShouldBeTrue)
	})
}

func TestInfoService_Info(t *testing.T) {
	convey.Convey("获取版本", t, func() {
		version, err := testClient.Info.Info()
		if err != nil {
			t.Error(err)
		}
		convey.So(version.Version, convey.ShouldNotBeEmpty)
	})
}
