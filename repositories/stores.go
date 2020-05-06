package repositories

//Store struct
type Store struct {
	ID        string
	StoreId   string `bson:"store_id"`
	Name      string
	HugeValue string
}
