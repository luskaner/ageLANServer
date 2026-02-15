package models

import (
	"encoding/json"
	"iter"
	"maps"
	"slices"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	i "github.com/luskaner/ageLANServer/server/internal"
)

type ItemCategoryRaw struct {
	Id            int32                 `json:"categoryID"`
	Name          string                `json:"name"`
	Group         string                `json:"categoryGroup"`
	Metadata      *ReadOnlyItemMetadata `json:"metadata"`
	LocalizedName string                `json:"localizedName,omitempty"`
}

type ItemDefinitionRaw struct {
	Id                   int32                 `json:"id"`
	Name                 string                `json:"name"`
	ImageData            string                `json:"imageData"`
	MetaData             *ReadOnlyItemMetadata `json:"metaData"`
	Available            int32                 `json:"available"`
	Version              int32                 `json:"version"`
	Level                int32                 `json:"level"`
	CategoryIDs          []int32               `json:"categoryIDs"`
	LocalizedName        string                `json:"localizedName,omitempty"`
	LocalizedDescription string                `json:"localizedDescription,omitempty"`
}

type ItemsRaw struct {
	Categories  []ItemCategoryRaw   `json:"itemCategories"`
	Definitions []ItemDefinitionRaw `json:"itemDefinitions"`
}

type ItemCategory interface {
	GetId() int32
	GetName() string
	GetGroup() string
	GetMetadata() *ReadOnlyItemMetadata
	GetLocalizedName() string
}

type MainItemCategory struct {
	id            int32
	name          string
	group         string
	metadata      *ReadOnlyItemMetadata
	localizedName string
}

func (c *MainItemCategory) GetId() int32 {
	return c.id
}

func (c *MainItemCategory) GetName() string {
	return c.name
}

func (c *MainItemCategory) GetGroup() string {
	return c.group
}

func (c *MainItemCategory) GetMetadata() *ReadOnlyItemMetadata {
	return c.metadata
}

func (c *MainItemCategory) GetLocalizedName() string {
	return c.localizedName
}

type Attribute struct {
	Value  string `json:"val"`
	Source int    `json:"source,omitempty,omitzero"`
	Mtype  int    `json:"mtype,omitempty,omitzero"`
}

type ReadOnlyItemMetadata struct {
	attributes map[string]Attribute
	other      map[string]json.RawMessage
}

func (r *ReadOnlyItemMetadata) GetAttribute(key string) (Attribute, bool) {
	if r.attributes == nil {
		return Attribute{}, false
	}
	attr, ok := r.attributes[key]
	return attr, ok
}

func (r *ReadOnlyItemMetadata) UnmarshalJSON(data []byte) error {
	var all map[string]json.RawMessage
	var allStr string
	var allData []byte
	if err := json.Unmarshal(data, &allStr); err != nil {
		return err
	} else if allStr == "" {
		return nil
	} else {
		allData = []byte(allStr)
	}
	if err := json.Unmarshal(allData, &all); err != nil {
		return nil
	}
	if attributes, ok := all["att"]; ok {
		if err := json.Unmarshal(attributes, &r.attributes); err != nil {
			return err
		}
		delete(all, "att")
	}
	if len(all) > 0 {
		r.other = maps.Clone(all)
	}
	return nil
}

type ItemMetadata struct {
	attributes map[string]string
	other      map[string]any
}

func (i *ItemMetadata) MarshalJSON() ([]byte, error) {
	all := make(map[string]any)
	if len(i.attributes) > 0 {
		all["attributes"] = i.attributes
	}
	if len(i.other) > 0 {
		all["other"] = maps.Clone(i.other)
	}
	return json.Marshal(all)
}

func (i *ItemMetadata) UnmarshalJSON(data []byte) error {
	var all map[string]json.RawMessage
	if err := json.Unmarshal(data, &all); err != nil {
		return err
	}
	if attributes, ok := all["attributes"]; ok {
		if err := json.Unmarshal(attributes, &i.attributes); err != nil {
			return err
		}
	}
	if other, ok := all["other"]; ok {
		if err := json.Unmarshal(other, &i.other); err != nil {
			return err
		}
	}
	return nil
}

func (i *ItemMetadata) Encode() (result string) {
	all := make(map[string]any)
	if len(i.attributes) > 0 {
		all["att"] = i.attributes
	}
	for k, v := range i.other {
		all[k] = v
	}
	if res, err := json.Marshal(all); err == nil {
		result = string(res)
	}
	return
}

func (i *ItemMetadata) GetAttribute(key string) (string, bool) {
	attr, ok := i.attributes[key]
	return attr, ok
}

func (i *ItemMetadata) UpdateAttribute(key string, value string) {
	i.attributes[key] = value
}

func (i *ItemMetadata) UpdateOther(key string, value any) {
	i.other[key] = value
}

type ItemDefinition interface {
	GetId() int32
	GetName() string
	GetMetadata() *ReadOnlyItemMetadata
	GetAvailable() int32
	GetVersion() int32
	GetLevel() int32
	GetCategories() mapset.Set[ItemCategory]
	GetLocalizedName() string
	GetLocalizedDescription() string
}

type MainItemDefinition struct {
	id                   int32
	name                 string
	metadata             *ReadOnlyItemMetadata
	available            int32
	version              int32
	level                int32
	categoryIDs          mapset.Set[ItemCategory]
	localizedName        string
	localizedDescription string
}

func (d *MainItemDefinition) GetId() int32 {
	return d.id
}

func (d *MainItemDefinition) GetName() string {
	return d.name
}

func (d *MainItemDefinition) GetMetadata() *ReadOnlyItemMetadata {
	return d.metadata
}

func (d *MainItemDefinition) GetAvailable() int32 {
	return d.available
}

func (d *MainItemDefinition) GetVersion() int32 {
	return d.version
}

func (d *MainItemDefinition) GetLevel() int32 {
	return d.level
}

func (d *MainItemDefinition) GetCategories() mapset.Set[ItemCategory] {
	return d.categoryIDs
}

func (d *MainItemDefinition) GetLocalizedName() string {
	return d.localizedName
}

func (d *MainItemDefinition) GetLocalizedDescription() string {
	return d.localizedDescription
}

type Item interface {
	Encode(userId int32) i.A
	GetId() int32
	SetLocationId(locationId int32)
	GetMetadata() *ItemMetadata
	IncrementVersion()
	SetDurabilityCount(charges int32)
}

type itemStorage struct {
	Id              int32
	Metadata        ItemMetadata `json:",omitempty,omitempty"`
	DefinitionId    int32
	DurabilityId    int32     `json:",omitempty,omitempty"`
	DurabilityCount int32     `json:",omitempty,omitempty"`
	CreationDate    time.Time `json:",omitempty,omitempty"`
	LocationId      int32     `json:",omitempty,omitempty"`
	TradeId         int32     `json:",omitempty,omitempty"`
	PermissionFlags uint32    `json:",omitempty,omitempty"`
	MaxCharges      int32     `json:",omitempty,omitempty"`
}

type MainItem struct {
	id              int32
	metadata        ItemMetadata
	entityVersion   int32
	definitionId    int32
	durabilityId    int32
	durabilityCount int32
	creationDate    time.Time
	locationId      int32
	tradeId         int32
	permissionFlags uint32
	maxCharges      int32
}

func newMainItemFromStorage(storage *itemStorage) *MainItem {
	return &MainItem{
		id:              storage.Id,
		metadata:        storage.Metadata,
		definitionId:    storage.DefinitionId,
		durabilityId:    storage.DurabilityId,
		durabilityCount: storage.DurabilityCount,
		creationDate:    storage.CreationDate,
		locationId:      storage.LocationId,
		tradeId:         storage.TradeId,
		permissionFlags: storage.PermissionFlags,
		maxCharges:      storage.MaxCharges,
	}
}

func (item *MainItem) MarshalJSON() ([]byte, error) {
	storage := &itemStorage{
		Id:              item.id,
		Metadata:        item.metadata,
		DefinitionId:    item.definitionId,
		DurabilityId:    item.durabilityId,
		DurabilityCount: item.durabilityCount,
		CreationDate:    item.creationDate,
		LocationId:      item.locationId,
		TradeId:         item.tradeId,
		PermissionFlags: item.permissionFlags,
		MaxCharges:      item.maxCharges,
	}
	return json.Marshal(storage)
}

func (item *MainItem) UnmarshalJSON(data []byte) error {
	var storage itemStorage
	if err := json.Unmarshal(data, &storage); err != nil {
		return err
	}
	*item = *newMainItemFromStorage(&storage)
	return nil
}

func (item *MainItem) GetId() int32 {
	return item.id
}

func (item *MainItem) SetLocationId(locationId int32) {
	item.locationId = locationId
}

func (item *MainItem) Encode(userId int32) i.A {
	return i.A{
		item.id,
		item.entityVersion,
		item.definitionId,
		userId,
		item.durabilityCount,
		item.durabilityId,
		item.metadata.Encode(),
		item.creationDate.Unix(),
		item.locationId,
		item.tradeId,
		item.permissionFlags,
		item.maxCharges,
	}
}

func (item *MainItem) GetMetadata() *ItemMetadata {
	return &item.metadata
}

func (item *MainItem) SetDurabilityCount(durabilityCount int32) {
	item.durabilityCount = durabilityCount
}

func (item *MainItem) IncrementVersion() {
	item.entityVersion++
}

type Items interface {
	GetLocation(id int32) (ItemLocation, bool)
	EncodeLocations() i.A
	GetCategory(id int32) (ItemCategory, bool)
	IterCategories() iter.Seq[ItemCategory]
	GetDefinition(id int32) (ItemDefinition, bool)
	IterDefinitions() iter.Seq[ItemDefinition]
	Initialize(itemDefinitions []byte, itemLocations i.A)
}

type ItemLocation interface {
	GetUnknown1IdxAutoNumeric() int32
	GetId() int32
	GetCategory() ItemCategory
	// GetMax is unverified
	GetMax() int32
	GetUnknown1() bool
	GetUnknown2() bool
	GetUnknown3() bool
	GetUnknown4() bool
	Encode() i.A
}

type MainItemLocation struct {
	unknown1IdxAutoNumeric int32
	id                     int32
	category               ItemCategory
	max                    int32
	unknown1               *i.BoolMappedNumber[int32]
	unknown2               *i.BoolMappedNumber[int32]
	unknown3               *i.BoolMappedNumber[int32]
	unknown4               *i.BoolMappedNumber[int32]
}

func (loc *MainItemLocation) GetId() int32 {
	return loc.id
}

func (loc *MainItemLocation) GetCategory() ItemCategory {
	return loc.category
}

func (loc *MainItemLocation) GetMax() int32 {
	return loc.max
}

func (loc *MainItemLocation) GetUnknown1IdxAutoNumeric() int32 {
	return loc.unknown1IdxAutoNumeric
}

func (loc *MainItemLocation) GetUnknown1() bool {
	return loc.unknown1.Bool()
}

func (loc *MainItemLocation) GetUnknown2() bool {
	return loc.unknown2.Bool()
}

func (loc *MainItemLocation) GetUnknown3() bool {
	return loc.unknown3.Bool()
}

func (loc *MainItemLocation) GetUnknown4() bool {
	return loc.unknown4.Bool()
}

func (loc *MainItemLocation) Encode() i.A {
	return i.A{
		loc.unknown1IdxAutoNumeric,
		0,
		loc.id,
		loc.category.GetId(),
		loc.max,
		loc.unknown1,
		loc.unknown2,
		loc.unknown3,
		loc.unknown4,
	}
}

type MainItems struct {
	categories  map[int32]ItemCategory
	definitions map[int32]ItemDefinition
	locations   map[int32]ItemLocation
}

func (m *MainItems) IterDefinitions() iter.Seq[ItemDefinition] {
	return maps.Values(m.definitions)
}

func (m *MainItems) IterCategories() iter.Seq[ItemCategory] {
	return maps.Values(m.categories)
}

func (m *MainItems) EncodeLocations() i.A {
	result := make(i.A, len(m.locations))
	for j, location := range m.locations {
		result[j] = location.Encode()
	}
	return result
}

func (m *MainItems) GetCategory(id int32) (ItemCategory, bool) {
	val, ok := m.categories[id]
	return val, ok
}

func (m *MainItems) GetDefinition(id int32) (ItemDefinition, bool) {
	val, ok := m.definitions[id]
	return val, ok
}

func (m *MainItems) GetLocation(id int32) (ItemLocation, bool) {
	val, ok := m.locations[id]
	return val, ok
}

func (m *MainItems) Initialize(itemDefinitions []byte, itemLocations i.A) {
	var definitions ItemsRaw
	if err := json.Unmarshal(itemDefinitions, &definitions); err != nil {
		return
	}
	m.categories = make(map[int32]ItemCategory)
	for _, categoryRaw := range definitions.Categories {
		metadata := categoryRaw.Metadata
		if metadata != nil && len(metadata.attributes) == 0 && len(metadata.other) == 0 {
			metadata = nil
		}
		category := &MainItemCategory{
			id:            categoryRaw.Id,
			name:          categoryRaw.Name,
			group:         categoryRaw.Group,
			metadata:      metadata,
			localizedName: categoryRaw.LocalizedName,
		}
		m.categories[category.id] = category
	}
	m.definitions = make(map[int32]ItemDefinition)
	for _, definitionRaw := range definitions.Definitions {
		categoryIDs := mapset.NewSet[ItemCategory]()
		for _, categoryID := range definitionRaw.CategoryIDs {
			if category, ok := m.categories[categoryID]; ok {
				categoryIDs.Add(category)
			}
		}
		metadata := definitionRaw.MetaData
		if metadata != nil && len(metadata.attributes) == 0 && len(metadata.other) == 0 {
			metadata = nil
		}
		definition := &MainItemDefinition{
			id:                   definitionRaw.Id,
			name:                 definitionRaw.Name,
			metadata:             metadata,
			available:            definitionRaw.Available,
			version:              definitionRaw.Version,
			level:                definitionRaw.Level,
			categoryIDs:          categoryIDs,
			localizedName:        definitionRaw.LocalizedName,
			localizedDescription: definitionRaw.LocalizedDescription,
		}
		m.definitions[definition.id] = definition
	}
	m.locations = make(map[int32]ItemLocation)
	for _, locationRawAny := range itemLocations {
		locationRawArr := locationRawAny.(i.A)
		location := &MainItemLocation{
			unknown1IdxAutoNumeric: int32(locationRawArr[0].(float64)),
			id:                     int32(locationRawArr[2].(float64)),
			max:                    int32(locationRawArr[4].(float64)),
			unknown1:               i.NewBoolMappedNumber(int32(locationRawArr[5].(float64))),
			unknown2:               i.NewBoolMappedNumber(int32(locationRawArr[6].(float64))),
			unknown3:               i.NewBoolMappedNumber(int32(locationRawArr[7].(float64))),
			unknown4:               i.NewBoolMappedNumber(int32(locationRawArr[8].(float64))),
		}
		if category, ok := m.categories[int32(locationRawArr[3].(float64))]; ok {
			location.category = category
		}
		m.locations[location.id] = location
	}
}

type ReadOnlyCategories struct {
	fromId map[int]ItemCategory
}

func (c *ReadOnlyCategories) GetById(id int) (ItemCategory, bool) {
	itemCategory, ok := c.fromId[id]
	return itemCategory, ok
}

type ReadOnlyItemDefinitions struct {
	fromId   map[int]ItemDefinition
	fromName map[string]ItemDefinition
}

func (d *ReadOnlyItemDefinitions) GetById(id int) (ItemDefinition, bool) {
	itemDefinition, ok := d.fromId[id]
	return itemDefinition, ok
}

func (d *ReadOnlyItemDefinitions) GetByName(name string) (ItemDefinition, bool) {
	itemDefinition, ok := d.fromName[name]
	return itemDefinition, ok
}

type ItemsUpgradableDefaultData struct {
	InitialUpgradableDefaultData[*map[int32]Item]
	gameId      string
	definitions Items
}

func NewItemsUpgradableDefaultData(gameId string, definitions Items) *ItemsUpgradableDefaultData {
	return &ItemsUpgradableDefaultData{
		InitialUpgradableDefaultData: InitialUpgradableDefaultData[*map[int32]Item]{},
		gameId:                       gameId,
		definitions:                  definitions,
	}
}

func (is *ItemsUpgradableDefaultData) Default() *map[int32]Item {
	var items []Item
	var itemPackCategory ItemCategory
	for category := range is.definitions.IterCategories() {
		if category.GetName() == "ItemPack" {
			itemPackCategory = category
			break
		}
	}
	for itemDefinition := range is.definitions.IterDefinitions() {
		if itemPackCategory != nil && itemDefinition.GetCategories().ContainsOne(itemPackCategory) {
			continue
		}
		var itemId int32
		i.WithRng(func(rand *i.RandReader) {
			for itemId = rand.Int32(); itemId < 100 || slices.ContainsFunc(items, func(item Item) bool {
				return item.GetId() == itemId
			}); {
			}
		})
		metadata := ItemMetadata{}
		if md := itemDefinition.GetMetadata(); md != nil {
			if attrs := md.attributes; len(attrs) > 0 {
				metadata.attributes = make(map[string]string, len(attrs))
				for k, v := range attrs {
					metadata.attributes[k] = v.Value
				}
			}
		}
		items = append(items, &MainItem{
			id:              itemId,
			metadata:        metadata,
			definitionId:    itemDefinition.GetId(),
			durabilityCount: 1,
			creationDate:    time.Now().UTC(),
			// Either another item id or location id as defined in statically per game
			tradeId: -1,
			// Max value seen given it's a binary flag
			permissionFlags: 63,
			maxCharges:      -1,
		})
	}
	itemsMap := make(map[int32]Item)
	for _, item := range items {
		itemsMap[item.GetId()] = item
	}
	return &itemsMap
}
