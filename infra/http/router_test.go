package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/minghsu0107/saga-purchase/infra/http/middleware"
	"github.com/minghsu0107/saga-purchase/infra/http/presenter"
	mock_repo "github.com/minghsu0107/saga-purchase/mock/repo"

	"github.com/golang/mock/gomock"
	conf "github.com/minghsu0107/saga-purchase/config"
	"github.com/minghsu0107/saga-purchase/domain/model"
	"github.com/minghsu0107/saga-purchase/infra/broker"
	mock_service "github.com/minghsu0107/saga-purchase/mock/service"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

var (
	mockCtrl              *gomock.Controller
	mockSubscriber        *MockSubscriber
	mockAuthRepo          *mock_repo.MockAuthRepository
	mockPurchaseResultSvc *mock_service.MockPurchaseResultService
	mockPurchasingSvc     *mock_service.MockPurchasingService
	server                *Server
)

func TestRouter(t *testing.T) {
	mockCtrl = gomock.NewController(t)
	RegisterFailHandler(Fail)
	RunSpecs(t, "router suite")
}

type MockSubscriber struct{}

type CustomClaims struct {
	CustomerID uint64
	jwt.StandardClaims
}

func (ms *MockSubscriber) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	outputChannel := make(chan *message.Message)
	return outputChannel, nil
}

func (ms *MockSubscriber) Close() error {
	return nil
}

func InitMocks() {
	mockAuthRepo = mock_repo.NewMockAuthRepository(mockCtrl)
	mockPurchasingSvc = mock_service.NewMockPurchasingService(mockCtrl)
}

func NewTestServer() *Server {
	config := &conf.Config{
		Logger: &conf.Logger{
			Writer: ioutil.Discard,
			ContextLogger: log.WithFields(log.Fields{
				"app": "test",
			}),
		},
	}
	log.SetOutput(config.Logger.Writer)
	engine := NewEngine(config)
	purchaseResultStreamHandler := NewPurchaseResultStreamHandler(mockPurchaseResultSvc)
	purchasingHandler := NewPurchasingHandler(mockPurchasingSvc)
	router := NewRouter(purchaseResultStreamHandler, purchasingHandler)
	sseRouter, _ := broker.NewSSERouter(mockSubscriber)
	jwtAuthChecker := middleware.NewJWTAuthChecker(config, mockAuthRepo)
	server := NewServer(config, engine, router, sseRouter, jwtAuthChecker)
	server.RegisterRoutes()
	return server
}

func GetResponse(router *gin.Engine, method, url string, body io.Reader) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, url, body)
	router.ServeHTTP(w, r)
	return w
}

func GetResponseWithBearerToken(router *gin.Engine, method, token, url string, body io.Reader) *httptest.ResponseRecorder {
	bearer := "Bearer " + token
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, url, body)
	r.Header.Set("Authorization", bearer)
	r.Header.Add("Accept", "application/json")
	//r.Header.Set("Accept", "text/event-stream") // this will block the server
	router.ServeHTTP(w, r)
	return w
}

func GetJSON(w *httptest.ResponseRecorder, target interface{}) error {
	body := ioutil.NopCloser(w.Body)
	defer body.Close()
	return json.NewDecoder(w.Body).Decode(target)
}

var _ = BeforeSuite(func() {
	InitMocks()
	server = NewTestServer()
})

var _ = AfterSuite(func() {
	mockCtrl.Finish()
})

var _ = Describe("router", func() {
	var purchasingEndpoint string
	var purchaseResultEndpoint string
	BeforeEach(func() {
		purchasingEndpoint = "/api/purchase"
		purchaseResultEndpoint = "/api/purchase/result"
	})
	Describe("test unauthorized request", func() {
		It("should return 401 unauthorized when trying to create purchase", func() {
			w := GetResponse(server.Engine, "POST", purchasingEndpoint, nil)
			Expect(w.Code).To(Equal(401))
		})
		It("should return 401 unauthorized when trying to retreive purchase result", func() {
			w := GetResponse(server.Engine, "GET", purchaseResultEndpoint, nil)
			Expect(w.Code).To(Equal(401))
		})
	})
	Describe("test access token", func() {
		var customerID uint64
		var tokenString string
		BeforeEach(func() {
			customerID = 1

			maxAge := 60 * 60 * 24
			customClaims := &CustomClaims{
				CustomerID: customerID,
				StandardClaims: jwt.StandardClaims{
					ExpiresAt: time.Now().Add(time.Duration(maxAge) * time.Second).Unix(),
					Issuer:    "testserver",
				},
			}
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, customClaims)
			secretKey := "243223ffslsfsldfl412fdsfsdf"
			tokenString, _ = token.SignedString([]byte(secretKey))
		})
		Describe("creating new purchase", func() {
			var testPurchase presenter.Purchase
			var body io.Reader
			BeforeEach(func() {
				testPurchase = presenter.Purchase{
					CartItems: &[]presenter.CartItem{
						{
							ProductID: 1,
							Amount:    3,
						},
					},
					Payment: &presenter.Payment{
						CurrencyCode: "NT",
					},
				}

				jsonBody, _ := json.Marshal(testPurchase)
				body = bytes.NewBuffer(jsonBody)
			})
			It("should success when passing valid access token", func() {
				mockAuthRepo.EXPECT().
					Auth(context.Background(), tokenString).Return(&model.AuthResult{
					CustomerID: customerID,
					Expired:    false,
				}, nil)
				mockPurchasingSvc.EXPECT().
					CreatePurchase(gomock.Any(), customerID, &testPurchase).Return(nil)
				w := GetResponseWithBearerToken(server.Engine, "POST", tokenString, purchasingEndpoint, body)
				Expect(w.Code).To(Equal(201))
			})
			It("should fail if using wrong method", func() {
				w := GetResponseWithBearerToken(server.Engine, "GET", tokenString, purchasingEndpoint, body)
				Expect(w.Code).To(Equal(404))

			})
			It("should fail if access token is not valid", func() {
				Context("when access token expired", func() {
					mockAuthRepo.EXPECT().
						Auth(context.Background(), tokenString).Return(&model.AuthResult{
						CustomerID: customerID,
						Expired:    true,
					}, nil)
					w := GetResponseWithBearerToken(server.Engine, "POST", tokenString, purchasingEndpoint, body)
					Expect(w.Code).To(Equal(401))
				})
				Context("when access token is invalid", func() {
					mockAuthRepo.EXPECT().
						Auth(context.Background(), tokenString).Return(nil, errors.New("invalid token"))
					w := GetResponseWithBearerToken(server.Engine, "POST", tokenString, purchasingEndpoint, body)
					Expect(w.Code).To(Equal(401))
				})
				Context("when access token is empty", func() {
					w := GetResponseWithBearerToken(server.Engine, "POST", "", purchasingEndpoint, body)
					Expect(w.Code).To(Equal(401))
				})
			})
		})
		Describe("before streaming purchase result", func() {
			It("should receive empty object if there is no purchase result", func() {
				mockAuthRepo.EXPECT().
					Auth(context.Background(), tokenString).Return(&model.AuthResult{
					CustomerID: customerID,
					Expired:    false,
				}, nil)

				w := GetResponseWithBearerToken(server.Engine, "GET", tokenString, purchaseResultEndpoint, nil)
				Expect(w.Code).To(Equal(200))
				receivedPurchaseResult := &presenter.PurchaseResult{}
				GetJSON(w, receivedPurchaseResult)
				Expect(receivedPurchaseResult).To(Equal(&presenter.PurchaseResult{}))
			})
			It("should fail if using wrong method", func() {
				w := GetResponseWithBearerToken(server.Engine, "POST", tokenString, purchaseResultEndpoint, nil)
				Expect(w.Code).To(Equal(404))

			})
			It("should fail if access token is not valid", func() {
				Context("when access token expired", func() {
					mockAuthRepo.EXPECT().
						Auth(context.Background(), tokenString).Return(&model.AuthResult{
						CustomerID: customerID,
						Expired:    true,
					}, nil)
					w := GetResponseWithBearerToken(server.Engine, "GET", tokenString, purchaseResultEndpoint, nil)
					Expect(w.Code).To(Equal(401))
				})
				Context("when access token is invalid", func() {
					mockAuthRepo.EXPECT().
						Auth(context.Background(), tokenString).Return(nil, errors.New("invalid token"))
					w := GetResponseWithBearerToken(server.Engine, "GET", tokenString, purchaseResultEndpoint, nil)
					Expect(w.Code).To(Equal(401))
				})
				Context("when access token is empty", func() {
					w := GetResponseWithBearerToken(server.Engine, "GET", "", purchaseResultEndpoint, nil)
					Expect(w.Code).To(Equal(401))
				})
			})
		})
	})
})
