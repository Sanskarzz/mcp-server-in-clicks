package storage

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStore struct {
	client *mongo.Client
	db     *mongo.Database
}

func NewMongoStore(ctx context.Context, uri, db string) (*MongoStore, error) {
	cli, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	// Try ping but allow startup without a live Mongo (useful for dev)
	ctx2, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	_ = cli.Ping(ctx2, nil)
	return &MongoStore{client: cli, db: cli.Database(db)}, nil
}

func (m *MongoStore) Close(ctx context.Context) error { return m.client.Disconnect(ctx) }

// Models
type ServerDef struct {
	ID         string                 `bson:"_id,omitempty" json:"id"`
	OwnerID    string                 `bson:"owner_id" json:"owner_id"`
	Name       string                 `bson:"name" json:"name"`
	ConfigJSON map[string]interface{} `bson:"config_json" json:"config_json"`
	CreatedAt  time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time              `bson:"updated_at" json:"updated_at"`
}

func (m *MongoStore) Servers() *mongo.Collection { return m.db.Collection("servers") }

// Multi-tenant models
type User struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	Email     string    `bson:"email" json:"email"`
	Name      string    `bson:"name" json:"name"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}

type Workspace struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	TenantID  string    `bson:"tenant_id" json:"tenant_id"`
	Name      string    `bson:"name" json:"name"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}

type Membership struct {
	ID          string    `bson:"_id,omitempty" json:"id"`
	WorkspaceID string    `bson:"workspace_id" json:"workspace_id"`
	UserID      string    `bson:"user_id" json:"user_id"`
	Role        string    `bson:"role" json:"role"` // owner|admin|member|guest
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
}

func (m *MongoStore) Users() *mongo.Collection       { return m.db.Collection("users") }
func (m *MongoStore) Workspaces() *mongo.Collection  { return m.db.Collection("workspaces") }
func (m *MongoStore) Memberships() *mongo.Collection { return m.db.Collection("memberships") }
