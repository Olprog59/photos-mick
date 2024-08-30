package models

// CollectionStats contient les statistiques de la collection de médias
type CollectionStats struct {
	TotalMedias int
	PhotoCount  int
	VideoCount  int
	OldestDate  string
	NewestDate  string
}
