package models

type ShopItemType string

const (
	ShopItemTypeFashionableBowl ShopItemType = "FASHIONABLE_BOWL"
	ShopItemTypeCyberBowl       ShopItemType = "CYBER_BOWL"
	ShopItemTypeHelperBowl      ShopItemType = "HELPER_BOWL"
	ShopItemTypeTraderBed       ShopItemType = "TRADER_BED"
	ShopItemTypeAccidentFreeBed ShopItemType = "ACCIDENT_FREE_BED"
	ShopItemTypeProBed          ShopItemType = "PRO_BED"
)

type ShopItem struct {
	ID            string
	Type          ShopItemType
	Title         string
	Description   string
	Category      RewardCategory
	RequiredLevel int
	PriceLeaves   int64
	DurationDays  int
}
