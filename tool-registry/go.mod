module github.com/blcvn/backend/services/ba-tool-registry

go 1.24.0

require (
	google.golang.org/grpc v1.62.1
	google.golang.org/protobuf v1.33.0
	gorm.io/driver/postgres v1.5.7
	gorm.io/gorm v1.25.10
)

replace github.com/blcvn/ba-shared-libs/pkg => ../../ba-shared-libs/pkg

replace github.com/blcvn/ba-shared-libs/proto => ../../ba-shared-libs/proto
