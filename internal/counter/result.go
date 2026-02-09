package counter

// LOCResult holds the aggregate results of counting lines of code in a repository.
type LOCResult struct {
	TotalLines    int64
	TotalCode     int64
	TotalComments int64
	TotalBlanks   int64
	TotalFiles    int64
	Languages     []LanguageStats
}

// LanguageStats holds LOC statistics for a single language.
type LanguageStats struct {
	Language string
	Lines    int64
	Code     int64
	Comments int64
	Blanks   int64
	Files    int64
}
