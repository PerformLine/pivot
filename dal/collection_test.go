package dal

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testGroup struct {
	ID   int    `pivot:"id,identity"`
	Name string `pivot:"name"`
}

type testUser struct {
	ID    int        `pivot:"id,identity"`
	Name  string     `pivot:"name"`
	Group *testGroup `pivot:"group_id"`
	Age   int        `pivot:"age"`
}

type nullBackend struct {
	collections map[string]*Collection
}

func (self *nullBackend) RegisterCollection(def *Collection) {
	if self.collections == nil {
		self.collections = make(map[string]*Collection)
	}

	def.SetBackend(self)
	self.collections[def.Name] = def
}

func (self *nullBackend) GetCollection(name string) (*Collection, error) {
	if c, ok := self.collections[name]; ok {
		return c, nil
	} else {
		return nil, CollectionNotFound
	}
}

func TestCollectionMakeRecord(t *testing.T) {
	assert := require.New(t)

	collection := NewCollection(`TestCollectionMakeRecord`)
	collection.AddFields([]Field{
		{
			Name: `name`,
			Type: StringType,
		}, {
			Name: `enabled`,
			Type: BooleanType,
		}, {
			Name: `age`,
			Type: IntType,
		},
	}...)

	type TestRecord struct {
		Name      string `pivot:"name"`
		Enabled   bool   `pivot:"enabled,omitempty"`
		ActualAge int    `pivot:"age"`
		Age       int
	}

	testRecord := TestRecord{
		Name:      `tester`,
		ActualAge: 42,
	}

	record, err := collection.MakeRecord(&testRecord)
	assert.Nil(err)
	assert.NotNil(record)

	assert.Equal(2, len(record.Fields))

	assert.Nil(record.ID)
	assert.Equal(`tester`, record.Get(`name`))
	assert.EqualValues(42, record.Get(`age`))

	type TestRecord2 struct {
		ID        int
		Name      string `pivot:"name"`
		Enabled   bool   `pivot:"enabled,omitempty"`
		ActualAge int    `pivot:"age"`
		Age       int
	}

	testRecord2 := TestRecord2{
		ID:   11,
		Name: `tester`,
	}

	record, err = collection.MakeRecord(&testRecord2)
	assert.Nil(err)
	assert.NotNil(record)

	assert.Equal(2, len(record.Fields))

	assert.Equal(11, record.ID)
	assert.Equal(`tester`, record.Get(`name`))
	assert.EqualValues(0, record.Get(`age`))
}

func TestCollectionMakeRecordRelated(t *testing.T) {
	assert := require.New(t)

	groups := NewCollection(`TestCollectionMakeRecordGroups`, Field{
		Name: `name`,
		Type: StringType,
	})

	users := NewCollection(`TestCollectionMakeRecordUsers`, Field{
		Name: `name`,
		Type: StringType,
	}, Field{
		Name:      `group_id`,
		Type:      IntType,
		BelongsTo: groups,
	}, Field{
		Name: `age`,
		Type: IntType,
	})

	backend := new(nullBackend)
	backend.RegisterCollection(users)
	backend.RegisterCollection(groups)

	record, err := users.MakeRecord(&testUser{
		Name: `tester`,
		Group: &testGroup{
			ID: 5432,
		},
		Age: 42,
	})
	assert.Nil(err)
	assert.NotNil(record)

	assert.Zero(record.ID)
	assert.Equal(`tester`, record.Get(`name`))
	assert.EqualValues(5432, record.Get(`group_id`))
	assert.EqualValues(42, record.Get(`age`))
}

func TestCollectionNewInstance(t *testing.T) {
	assert := require.New(t)

	constantTimeFn := func() time.Time {
		return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
	}

	collection := NewCollection(`TestCollectionNewInstance`)
	collection.AddFields([]Field{
		{
			Name:         `name`,
			Type:         StringType,
			DefaultValue: `Bob`,
		}, {
			Name:         `enabled`,
			Type:         BooleanType,
			DefaultValue: true,
		}, {
			Name:         `age`,
			Type:         IntType,
			DefaultValue: []string{`WRONG TYPE`},
		}, {
			Name:         `created_at`,
			Type:         TimeType,
			DefaultValue: constantTimeFn,
		},
	}...)

	type TestRecord struct {
		Name      string    `pivot:"name"`
		Enabled   bool      `pivot:"enabled,omitempty"`
		Age       int       `pivot:"age"`
		CreatedAt time.Time `pivot:"created_at"`
	}

	collection.SetRecordType(TestRecord{})
	assert.True(reflect.DeepEqual(
		collection.recordType,
		reflect.TypeOf(TestRecord{}),
	))

	instanceI := collection.NewInstance()
	instance, ok := instanceI.(*TestRecord)
	assert.True(ok)

	assert.Equal(`Bob`, instance.Name)
	assert.True(instance.Enabled)
	assert.Zero(instance.Age)
	assert.Equal(constantTimeFn(), instance.CreatedAt)
}

func TestCollectionValidator(t *testing.T) {
	assert := require.New(t)

	collection := &Collection{
		Name:              `TestCollectionValidator`,
		IdentityFieldType: StringType,
		PreSaveValidator: func(record *Record) error {
			if fmt.Sprintf("%v", record.ID) == `two` {
				return fmt.Errorf("ID cannot be 'two' for reasons")
			}

			return nil
		},
	}

	assert.NoError(collection.ValidateRecord(NewRecord(`one`), PersistOperation))
	assert.Error(collection.ValidateRecord(NewRecord(`two`), PersistOperation))
	assert.NoError(collection.ValidateRecord(NewRecord(`three`), PersistOperation))
}

func TestCollectionMapFromRecord(t *testing.T) {
	assert := require.New(t)

	collection := NewCollection(`TestCollectionMakeRecord`)
	collection.IdentityFieldFormatter = func(id interface{}, op FieldOperation) (interface{}, error) {
		if record, ok := id.(*Record); ok {
			id = record.ID
		}

		return fmt.Sprintf("ID<%v>", id), nil
	}

	collection.AddFields([]Field{
		{
			Name: `name`,
			Type: StringType,
		}, {
			Name:         `enabled`,
			Type:         BooleanType,
			DefaultValue: true,
		}, {
			Name: `age`,
			Type: IntType,
		},
	}...)

	rv, err := collection.MapFromRecord(
		NewRecord(`test`).Set(`name`, `tester`).Set(`age`, 42),
	)

	assert.NoError(err)
	assert.EqualValues(map[string]interface{}{
		`id`:      `ID<test>`,
		`name`:    `tester`,
		`enabled`: true,
		`age`:     42,
	}, rv)
}

func TestCollectionTTLField(t *testing.T) {
	assert := require.New(t)

	collection := &Collection{
		Name:              `TestCollectionTTLField`,
		IdentityFieldType: StringType,
		TimeToLiveField:   `ttl`,
	}

	assert.Zero(collection.TTL(NewRecord(`test1`)))
	assert.Zero(collection.TTL(NewRecord(`test1`).Set(`ttl`, nil)))
	assert.Zero(collection.TTL(NewRecord(`test1`).Set(`ttl`, 0)))
	assert.Zero(collection.TTL(NewRecord(`test1`).Set(`ttl`, ``)))
	assert.Zero(collection.TTL(NewRecord(`test1`).Set(`ttl`, false)))

	assert.NotZero(
		collection.TTL(NewRecord(`test1`).Set(`ttl`, time.Now().Add(time.Second))),
	)

	assert.NotZero(
		collection.TTL(NewRecord(`test1`).Set(`ttl`, time.Now().Unix()+60)),
	)
}

func TestCollectionIsExpired(t *testing.T) {
	assert := require.New(t)

	collection := &Collection{
		Name:              `TestCollectionIsExpired`,
		IdentityFieldType: StringType,
		TimeToLiveField:   `ttl`,
	}

	assert.False(collection.IsExpired(NewRecord(`test1`)))
	assert.False(collection.IsExpired(NewRecord(`test1`).Set(`ttl`, nil)))
	assert.False(collection.IsExpired(NewRecord(`test1`).Set(`ttl`, 0)))
	assert.False(collection.IsExpired(NewRecord(`test1`).Set(`ttl`, ``)))
	assert.False(collection.IsExpired(NewRecord(`test1`).Set(`ttl`, false)))

	assert.False(
		collection.IsExpired(NewRecord(`test1`).Set(`ttl`, time.Now().Add(time.Second))),
	)

	assert.False(
		collection.IsExpired(NewRecord(`test1`).Set(`ttl`, time.Now().Unix()+60)),
	)

	assert.True(
		collection.IsExpired(NewRecord(`test1`).Set(`ttl`, time.Now().Add(-1*time.Nanosecond))),
	)

	assert.True(
		collection.IsExpired(NewRecord(`test1`).Set(`ttl`, time.Now().Unix()-60)),
	)
}

func TestCollectionExtractValueFromRelationship(t *testing.T) {
	assert := require.New(t)

	groups := NewCollection(`TestCollectionExtractValueFromRelationshipGroups`, Field{
		Name: `name`,
		Type: StringType,
	})

	users := NewCollection(`TestCollectionExtractValueFromRelationshipUsers`, Field{
		Name: `name`,
		Type: StringType,
	}, Field{
		Name:      `group_id`,
		Type:      IntType,
		BelongsTo: groups,
	}, Field{
		Name: `age`,
		Type: IntType,
	})

	backend := new(nullBackend)
	backend.RegisterCollection(users)
	backend.RegisterCollection(groups)

	// test ability for Users to extract key values from Groups
	f, ok := users.GetField(`group_id`)
	assert.True(ok)
	keys, err := users.extractValueFromRelationship(&f, &testGroup{
		ID: 5432,
	}, PersistOperation)

	assert.NoError(err)
	assert.EqualValues(5432, keys)

}
