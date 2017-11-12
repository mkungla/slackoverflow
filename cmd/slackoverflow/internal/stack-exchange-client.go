// Copyright © 2016 -2017 A-Frame authors.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package internal

import (
	"fmt"
	"net/url"

	"github.com/howi-ce/howi/addon/application/plugin/cli"
)

// StackExchangeClient to call Stack Exchange API
type StackExchangeClient struct {
	quotaMax       int
	quotaRemaining int
	apiHost        string
	apiVersion     string
	apiKey         string
}

// GetQuotaRemaining return remaining quota for today
func (s *StackExchangeClient) GetQuotaRemaining() int {
	return s.quotaRemaining
}

// GetQuotaMax return maximum allowed quota
func (s *StackExchangeClient) GetQuotaMax() int {
	return s.quotaMax
}

// SetQuotaRemaining set remaining quota for today
func (s *StackExchangeClient) SetQuotaRemaining(quotaRemaining int) {
	s.quotaRemaining = quotaRemaining
}

// SetQuotaMax set maximum allowed quota
func (s *StackExchangeClient) SetQuotaMax(quotaMax int) {
	s.quotaMax = quotaMax
}

// GetEndpont returns base endpoint
func (s *StackExchangeClient) GetEndpont(path string) (*url.URL, error) {
	return url.Parse(s.apiHost + "/" + s.apiVersion + "/" + path)
}

// SetHost set Stack Exchange API host
func (s *StackExchangeClient) SetHost(host string) {
	s.apiHost = host
}

// SetAPIVersion set Stack Exchange API version
func (s *StackExchangeClient) SetAPIVersion(varsion string) {
	s.apiVersion = varsion
}

// SetKey Set Stack Exchange API to receive a higher request quota if exists
func (s *StackExchangeClient) SetKey(apiKey string) {
	s.apiKey = apiKey
}

// SearchAdvanced https://api.stackexchange.com/docs/advanced-search
func (s *StackExchangeClient) SearchAdvanced() *SearchAdvanced {
	searchAdvanced := &SearchAdvanced{}
	searchAdvanced.Client = s
	searchAdvanced.Init()
	return searchAdvanced
}

// Questions https://api.stackexchange.com/docs/questions-by-ids
func (s *StackExchangeClient) Questions() *Questions {
	questions := &Questions{}
	questions.Client = s
	questions.Init()
	return questions
}

// SearchAdvanced - https://api.stackexchange.com/docs/advanced-search
type SearchAdvanced struct {
	Client     *StackExchangeClient
	Parameters Parameters
	Paging
	Result *QuestionsWrapperObj
}

// Init initializes Search Advanced module
func (sa *SearchAdvanced) Init() {
	sa.Parameters.Allow("site", "stackoverflow",
		"site where to check questions from")
	sa.Parameters.Allow("q", "",
		"a free form text parameter, will match all question properties based on an undocumented algorithm.")
	sa.Parameters.Allow("accepted", "",
		"true to return only questions with accepted answers, false to return only those without. Omit to elide constraint.")
	sa.Parameters.Allow("answers", "",
		"the minimum number of answers returned questions must have.")
	sa.Parameters.Allow("body", "",
		"text which must appear in returned questions' bodies.")
	sa.Parameters.Allow("closed", "",
		"true to return only closed questions, false to return only open ones. Omit to elide constraint.")
	sa.Parameters.Allow("migrated", "",
		"true to return only questions migrated away from a site, false to return only those not. Omit to elide constraint.")
	sa.Parameters.Allow("notice", "",
		"true to return only questions with post notices, false to return only those without. Omit to elide constraint.")
	sa.Parameters.Allow("nottagged", "",
		"a semicolon delimited list of tags, none of which will be present on returned questions.")
	sa.Parameters.Allow("tagged", "",
		"a semicolon delimited list of tags, of which at least one will be present on all returned questions.")
	sa.Parameters.Allow("title", "",
		"text which must appear in returned questions titles.")
	sa.Parameters.Allow("user", "",
		"the id of the user who must own the questions returned.")
	sa.Parameters.Allow("url", "",
		"a url which must be contained in a post, may include a wildcard.")
	sa.Parameters.Allow("views", "",
		"the minimum number of views returned questions must have.")
	sa.Parameters.Allow("wiki", "",
		"true to return only community wiki questions, false to return only non-community wiki ones. Omit to elide constraint.")
	sa.Parameters.Allow("sort", "creation",
		"The sorts accepted by this method operate on the follow fields of the question object: activity, creation, votes, relevance")
	sa.Parameters.Allow("order", "asc",
		"Order results is ascending or descending")
	sa.Parameters.Allow("fromdate", "",
		"From which date to search")
	sa.Parameters.Allow("todate", "",
		"Up to which date to search")
	sa.Parameters.Allow("filter", "!6hYwbNNZ(*eH3a3)XT0aZCOGTo-kwAtAoVF5vC378NPI6Y",
		"Defined Custom Filters https://api.stackexchange.com/docs/filters")
	sa.Parameters.Allow("pagesize", 100,
		"API. page starts at and defaults to 1, pagesize can be any value between 0 and 100")
	sa.Parameters.Allow("page", 1,
		"Current page to be fetched")
	sa.Parameters.Allow("key", sa.Client.apiKey,
		"Pass this as key when making requests against the Stack Exchange API to receive a higher request quota.")
}

// DrawQuery output table of search query
func (sa *SearchAdvanced) DrawQuery(w *cli.Worker) {
	w.Log.Line("Search Advanced query will be performed with following parameters")
	query := NewTable("parameter", "defined value", "default", "description")
	for key, allowed := range sa.Parameters.GetAllowed() {
		query.AddRow(key, sa.Parameters.ValueOf(key), allowed.String(), allowed.Decription())
	}
	query.Print()
	url, _ := sa.GetURL()
	w.Log.Linef("Resulting Url: %s", url)
}

// GetURL composed from current parameters
func (sa *SearchAdvanced) GetURL() (string, error) {
	endpoint, err := sa.Client.GetEndpont("search/advanced")

	query := endpoint.Query()
	// Apply default parameters
	sa.Parameters.ApplyDefaults()

	// Apply defined parameters
	for param, value := range sa.Parameters.GetApplied() {
		query.Set(param, value.String())
	}

	endpoint.RawQuery = query.Encode()

	return endpoint.String(), err
}

// Get request
func (sa *SearchAdvanced) Get() (bool, error) {

	url, err := sa.GetURL()

	if err != nil {
		return false, err
	}

	err = HTTPGetByURL(url, &sa.Result)
	if err != nil {
		fmt.Println(err.Error())
		return false, err
	}

	sa.Paging.curentPage = sa.Result.Page
	sa.Paging.hasMore = sa.Result.HasMore
	sa.Client.SetQuotaMax(sa.Result.QuotaMax)
	sa.Client.SetQuotaRemaining(sa.Result.QuotaRemaining)

	return true, err
}

// Questions - https://api.stackexchange.com/docs/questions-by-ids
type Questions struct {
	Client     *StackExchangeClient
	Parameters Parameters
	Paging
	Result *QuestionsWrapperObj
}

// Init initializes Questions module
func (q *Questions) Init() {
	q.Parameters.Allow("site", "stackoverflow",
		"site where to check questions from")

	q.Parameters.Allow("sort", "activity",
		"The sorts accepted by this method operate on the follow fields of the question object: activity, creation, votes")
	q.Parameters.Allow("order", "asc",
		"Order results is ascending or descending")
	q.Parameters.Allow("fromdate", "",
		"From which date to search")
	q.Parameters.Allow("todate", "",
		"Up to which date to search")

	q.Parameters.Allow("filter", "!6hYwbNNZ(*eH3a3)XT0aZCOGTo-kwAtAoVF5vC378NPI6Y",
		"Defined Custom Filters https://api.stackexchange.com/docs/filters")
	q.Parameters.Allow("pagesize", 100,
		"API. page starts at and defaults to 1, pagesize can be any value between 0 and 100")
	q.Parameters.Allow("page", 1,
		"Current page to be fetched")
	q.Parameters.Allow("min", "",
		"min and max specify the range of a field must fall in (that field being specified by sort)")
	q.Parameters.Allow("max", "",
		"min and max specify the range of a field must fall in (that field being specified by sort)")
	q.Parameters.Allow("key", q.Client.apiKey,
		"Pass this as key when making requests against the Stack Exchange API to receive a higher request quota.")
}

// DrawQuery output table of questions query
func (q *Questions) DrawQuery(w *cli.Worker, ids string) {
	w.Log.Line("Questions query will be performed with following parameters")
	query := NewTable("parameter", "defined value", "default", "description")
	for key, allowed := range q.Parameters.GetAllowed() {
		query.AddRow(key, q.Parameters.ValueOf(key), allowed.String(), allowed.Decription())
	}
	var idss string
	if len(ids) > 10 {
		idss = ids[0:10]
	} else {
		idss = ids
	}
	query.AddRow("ids", idss+"...", "",
		"{ids} can contain up to 100 semicolon delimited ids, to find ids programatically look for question_id ")
	query.Print()
	url, _ := q.GetURL(ids)
	w.Log.Linef("Resulting Url: %s", url)
}

// Get request
func (q *Questions) Get(ids string) (bool, error) {

	url, err := q.GetURL(ids)

	if err != nil {
		return false, err
	}

	err = HTTPGetByURL(url, &q.Result)
	if err != nil {
		fmt.Println(err.Error())
		return false, err
	}

	q.Paging.curentPage = q.Result.Page
	q.Paging.hasMore = q.Result.HasMore
	q.Client.SetQuotaMax(q.Result.QuotaMax)
	q.Client.SetQuotaRemaining(q.Result.QuotaRemaining)

	return true, err
}

// GetURL composed from current parameters
func (q *Questions) GetURL(ids string) (string, error) {
	// Apply default parameters
	q.Parameters.ApplyDefaults()

	// Make sure to set failing id if no ids are supplied
	if len(ids) == 0 {
		ids = "100"
	}

	endpoint, err := q.Client.GetEndpont("questions/" + ids)
	query := endpoint.Query()

	// Remove IDS since that is actually set in path
	q.Parameters.Delete("ids")

	// Apply defined parameters
	for param, value := range q.Parameters.GetApplied() {
		query.Set(param, value.String())
	}

	endpoint.RawQuery = query.Encode()

	return endpoint.String(), err
}

// QuestionsWrapperObj is a API response type
type QuestionsWrapperObj struct {
	Backoff        int           `json:"backoff"`
	ErrorID        int           `json:"error_id"`
	ErrorName      string        `json:"error_name"`
	ErrorMessage   string        `json:"error_message"`
	HasMore        bool          `json:"has_more"`
	Page           int           `json:"page"`
	QuotaMax       int           `json:"quota_max"`
	QuotaRemaining int           `json:"quota_remaining"`
	Items          []QuestionObj `json:"items"`
}

// QuestionObj is question item returned by StackExchange API
type QuestionObj struct {
	QID              int            `json:"question_id"`
	Title            string         `json:"title"`
	CreationDate     int64          `json:"creation_date"`
	LastActivityDate int64          `json:"last_activity_date"`
	Owner            ShallowUserObj `json:"owner"`
	IsAnswered       bool           `json:"is_answered"`
	ShareLink        string         `json:"share_link"`
	ClosedReason     string         `json:"closed_reason"`
	Tags             []string       `json:"tags"`
	Score            int            `json:"score"`
	ViewCount        int            `json:"view_count"`
	AnswerCount      int            `json:"answer_count"`
	CommentCount     int            `json:"comment_count"`
	UpVoteCount      int            `json:"up_vote_count"`
	DownVoteCount    int            `json:"down_vote_count"`
	DeleteVoteCount  int            `json:"delete_vote_count"`
	FavoriteCount    int            `json:"favorite_count"`
	ReOpenVoteCount  int            `json:"reopen_vote_count"`
}

// BadgeCountsObj of Stack Exchange User
type BadgeCountsObj struct {
	Bronze int `json:"bronze"`
	Silver int `json:"silver"`
	Gold   int `json:"gold"`
}

// ShallowUserObj returned by Stack Exchange API
type ShallowUserObj struct {
	UID          int            `json:"user_id"`
	Reputation   int            `json:"reputation"`
	ProfileImage string         `json:"profile_image"`
	DisplayName  string         `json:"display_name"`
	Link         string         `json:"link"`
	AcceptRate   int            `json:"accept_rate"`
	BadgeCounts  BadgeCountsObj `json:"badge_counts"`
}

// Parameters Stack Exchange query parameters
type Parameters struct {
	applied map[string]Parameter
	allowed map[string]Parameter
}

// Allow parameter to be set
func (p *Parameters) Allow(param string, value interface{}, desc string) {
	if p.allowed == nil {
		p.allowed = make(map[string]Parameter)
	}
	p.allowed[param] = Parameter{param, value, desc}
}

// IsSet return true is parameter is set
func (p *Parameters) IsSet(param string) bool {
	if _, ok := p.applied[param]; ok {
		return true
	}
	return false
}

// Set adds or moifies a parameter given by the key and value
func (p *Parameters) Set(param string, value interface{}) {
	if !p.IsAllowed(param) || value == "" {
		return
	}
	if p.applied == nil {
		p.applied = make(map[string]Parameter)
	}
	p.applied[param] = Parameter{param, value, ""}
}

// ValueOf returns value of given paramter if it is set
func (p *Parameters) ValueOf(param string) string {
	if val, ok := p.applied[param]; ok {
		return val.String()
	}
	return ""
}

// Delete the given parameter
func (p *Parameters) Delete(param string) {
	delete(p.applied, param)
}

// IsAllowed return true is parameter is allowed by this endpoint
func (p *Parameters) IsAllowed(param string) bool {
	if _, ok := p.allowed[param]; ok {
		return true
	}
	return false
}

// GetAllowed returns parameters accepted by this endpoint
func (p *Parameters) GetAllowed() map[string]Parameter {
	return p.allowed
}

// ApplyDefaults apply default parameters
func (p *Parameters) ApplyDefaults() {
	// Do nothing if no defaults are set
	if p.allowed == nil {
		return
	}
	for param, value := range p.GetAllowed() {
		// Set default value if parameter has not been set
		if !p.IsSet(param) {
			p.Set(param, value.String())
		}
	}
}

// GetApplied returns parameters which have been set
func (p *Parameters) GetApplied() map[string]Parameter {
	return p.applied
}

// Parameter Stack Exchange parameter
type Parameter struct {
	param       string
	value       interface{}
	description string
}

// String value of the parameter
func (p *Parameter) String() string {
	return fmt.Sprintf("%v", p.value)
}

// Decription of this parameter
func (p *Parameter) Decription() string {
	return p.description
}

// Paging - https://api.stackexchange.com/docs/paging
type Paging struct {
	page       int
	pageSize   int
	hasMore    bool
	curentPage int
}

// SetPage sets current page
func (p *Paging) SetPage(page int) {
	p.page = page
}

// GetCurrentPageNr return current page number
func (p *Paging) GetCurrentPageNr() int {
	if p.page == 0 {
		p.page = 1
	}
	return p.page
}

// SetPageSize sets current page
func (p *Paging) SetPageSize(pageSize int) {
	p.pageSize = pageSize
}

// NextPage increases page number
func (p *Paging) NextPage() {
	p.page++
}

// PreviousPage decreases page number
func (p *Paging) PreviousPage() {
	if p.page > 0 {
		p.page--
	}
}

// FirstPage sets current page to 1
func (p *Paging) FirstPage() {
	p.page = 1
}

// HasMore either there are more results
func (p *Paging) HasMore() bool {
	return p.hasMore
}
