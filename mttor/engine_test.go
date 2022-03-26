package mttor_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/teawithsand/arcah/mttor"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const MaxUint64 = ^uint64(0)
const MaxInt64 = int64(MaxUint64 >> 1)
const MinInt64 = -MaxInt64 - 1

type DataBson struct {
	Number int64  `bson:"numInBson"`
	Text   string `bson:"tExTiNbSoN"`
}

type Data struct {
	Number int64
	Text   string

	Ints []int
}

type DataSetText struct {
	Text string
}

type DataIncNumber struct {
	Number int64 `mttor:",inc"`
}

type DataCombined struct {
	Text   string `mttor:",,omitempty"`
	Number int64  `mttor:",inc,omitempty"`
}

type DataPushInt struct {
	Value int `mttor:"Ints,push"`
}

type DataPushInts struct {
	Values []int `mttor:"Ints,push"`
}

func TestMutator_OnObject(t *testing.T) {
	engine := mttor.NewDefaultMutatorEngine()
	t.Run("set", func(t *testing.T) {
		data := Data{}
		err := engine.Mutate(context.Background(), &data, &DataSetText{
			Text: "asdf",
		})
		if err != nil {
			t.Error(err)
			return
		}

		if data.Text != "asdf" {
			t.Error("data wasn't mutated")
			return
		}
	})

	t.Run("inc", func(t *testing.T) {
		data := Data{
			Number: 31,
		}
		err := engine.Mutate(context.Background(), &data, DataIncNumber{
			Number: 11,
		})
		if err != nil {
			t.Error(err)
			return
		}

		if data.Number != 42 {
			t.Error("data wasn't mutated", "got", data.Number)
			return
		}
	})

	t.Run("combined_set", func(t *testing.T) {
		data := Data{
			Number: 31,
			Text:   "fdsa",
		}
		err := engine.Mutate(context.Background(), &data, DataCombined{
			Number: 11,
		})
		if err != nil {
			t.Error(err)
			return
		}

		if data.Number != 42 {
			t.Error("data wasn't mutated", "got", data.Number)
			return
		}

		if data.Text != "fdsa" {
			t.Error("data was changed, while expected it not to")
			return
		}
	})

	t.Run("combined_inc", func(t *testing.T) {
		data := Data{
			Number: 31,
			Text:   "fdsa",
		}
		err := engine.Mutate(context.Background(), &data, DataCombined{
			Text: "asdf",
		})
		if err != nil {
			t.Error(err)
			return
		}

		if data.Text != "asdf" {
			t.Error("data wasn't mutated", "got", data.Text)
			return
		}

		if data.Number != 31 {
			t.Error("data was changed, while expected it not to")
			return
		}
	})
}

func DoTestMutationOnMongo(t *testing.T, engine mttor.MongoMutatorEngine, data, mutation interface{}) {
	uri := os.Getenv("ARCAH_TEST_MONGO")
	if len(uri) > 0 {
		client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
		if err != nil {
			t.Error(err)
			return
		}
		var buf [16]byte
		_, err = io.ReadFull(rand.Reader, buf[:])
		if err != nil {
			t.Error(err)
			return
		}
		dbName := hex.EncodeToString(buf[:])
		database := client.Database(dbName)
		collection := database.Collection("testdata")
		err = client.UseSession(context.Background(), func(ctx mongo.SessionContext) (err error) {
			_, err = collection.InsertOne(ctx, data)
			if err != nil {
				t.Error(err)
			}

			inDbTarget := reflect.New(reflect.TypeOf(data).Elem()).Interface()

			err = engine.Mutate(ctx, data, mutation)
			if err != nil {
				return
			}

			renderedMutation, err := engine.RenderMongoMutation(ctx, reflect.TypeOf(data), mutation)
			if err != nil {
				t.Error(err)
				return
			}

			_, err = collection.UpdateOne(ctx, bson.D{}, renderedMutation)
			if err != nil {
				t.Error(err)
				return
			}

			err = collection.FindOne(ctx, bson.D{}).Decode(inDbTarget)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(data, inDbTarget) {
				t.Error(fmt.Errorf("mutation mismatch in mongo and local one: expected %+#v got %+#v", data, inDbTarget))
				return
			}

			return
		})
		if err != nil {
			t.Error(err)
			return
		}

		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			database.Drop(ctx)
			client.Disconnect(ctx)
		}()
	}

}

func TestMutator_WithMongo(t *testing.T) {
	uri := os.Getenv("ARCAH_TEST_MONGO")
	if len(uri) == 0 {
		log.Default().Println("Note: omitting non-mongo tests")
		return
	}
	engine := mttor.NewDefaultMutatorEngine()

	t.Run("set", func(t *testing.T) {
		DoTestMutationOnMongo(t, engine, &Data{}, DataSetText{
			Text: "asdf",
		})
	})

	t.Run("inc", func(t *testing.T) {
		DoTestMutationOnMongo(t, engine, &Data{
			Number: 31,
		}, DataIncNumber{
			Number: 11,
		})
	})
	// Overflow case is not supported in mongo
	// it returns error,
	// wheres in go it's fine to do overflow
	// but I guess this difference is OK for most use cases
	/*
		t.Run("inc_overflow", func(t *testing.T) {
			DoTestMutationOnMongo(t, engine, &Data{
				Number: MaxInt64 - 1,
			}, DataIncNumber{
				Number: 1,
			})
		})
	*/

	t.Run("push", func(t *testing.T) {
		DoTestMutationOnMongo(t, engine, &Data{
			Ints: []int{1, 2, 3},
		}, DataPushInt{
			Value: 4,
		})
	})

	t.Run("push_many", func(t *testing.T) {
		DoTestMutationOnMongo(t, engine, &Data{
			Ints: []int{1, 2, 3},
		}, DataPushInts{
			Values: []int{4, 5, 6},
		})
	})
}
