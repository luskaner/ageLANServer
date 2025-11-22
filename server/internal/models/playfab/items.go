package playfab

const Iso8601Layout string = "2006-01-02T15:04:05.000Z"

type CatalogItemAlternativeId struct {
	Type  string
	Value string
}

type CatalogItemTitle struct {
	NEUTRAL string
}

type CatalogItemCreatorEntity struct {
	Id         string
	Type       string
	TypeString string
}

type CatalogItem struct {
	Id           string
	Type         string
	AlternateIds []CatalogItemAlternativeId
	FriendlyId   string
	Title        CatalogItemTitle
	Description  struct {
	}
	Keywords struct {
	}
	CreatorEntity     CatalogItemCreatorEntity
	Platforms         []any
	Tags              []any
	CreationDate      string
	LastModifiedDate  string
	StartDate         string
	Contents          []any
	Images            []any
	ItemReferences    []any
	DeepLinks         []any
	DisplayProperties struct {
	}
}

type InventoryItem struct {
	Id                string
	StackId           string
	DisplayProperties struct {
	}
	Amount int
	Type   string
}
