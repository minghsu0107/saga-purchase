package result

import (
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	conf "github.com/minghsu0107/saga-purchase/config"
	"github.com/minghsu0107/saga-purchase/domain/event"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

var purchaseResultSvc PurchaseResultService

func TestRouter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "result service suite")
}

var _ = BeforeSuite(func() {
	config := &conf.Config{
		Logger: &conf.Logger{
			Writer: ioutil.Discard,
			ContextLogger: log.WithFields(log.Fields{
				"app_name": "test",
			}),
		},
	}
	purchaseResultSvc = NewPurchaseResultService(config)
})

var _ = Describe("result service", func() {
	It("should pass purchase result to http request context correctly", func() {
		purchaseResult := &event.PurchaseResult{
			Step:   "dummpstep",
			Status: "dummystatus",
		}
		r := &http.Request{}
		r = r.WithContext(context.WithValue(r.Context(), conf.MsgKey, purchaseResult))
		curPurchaseResult, err := purchaseResultSvc.GetPurchaseResult(r)
		Expect(err).To(BeNil())
		Expect(curPurchaseResult).To(Equal(purchaseResult))
	})
})
