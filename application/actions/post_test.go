package actions

import (
	"github.com/silinternational/wecarry-api/models"
)

type PostResponse struct {
	Post struct {
		ID          string          `json:"id"`
		Type        models.PostType `json:"type"`
		Title       string          `json:"title"`
		Description string          `json:"description"`
		Destination struct {
			Description string  `json:"description"`
			Country     string  `json:"country"`
			Lat         float64 `json:"latitude"`
			Long        float64 `json:"longitude"`
		} `json:"destination"`
		Origin struct {
			Description string  `json:"description"`
			Country     string  `json:"country"`
			Lat         float64 `json:"latitude"`
			Long        float64 `json:"longitude"`
		} `json:"origin"`
		Size         models.PostSize   `json:"size"`
		NeededAfter  string            `json:"neededAfter"`
		NeededBefore string            `json:"neededBefore"`
		Category     string            `json:"category"`
		Status       models.PostStatus `json:"status"`
		CreatedAt    string            `json:"createdAt"`
		UpdatedAt    string            `json:"updatedAt"`
		Cost         string            `json:"cost"`
		IsEditable   bool              `json:"isEditable"`
		Url          string            `json:"url"`
		CreatedBy    struct {
			ID string `json:"id"`
		} `json:"createdBy"`
		Receiver struct {
			ID string `json:"id"`
		} `json:"receiver"`
		Provider struct {
			ID string `json:"id"`
		} `json:"provider"`
		Organization struct {
			ID string `json:"id"`
		} `json:"organization"`
		Photo struct {
			ID string `json:"id"`
		} `json:"photo"`
		Files []struct {
			ID string `json:"id"`
		} `json:"files"`
	} `json:"post"`
}

func (as *ActionSuite) Test_UpdatePost() {
	t := as.T()

	f := createFixturesForUpdatePost(as)

	var postsResp PostResponse

	input := `id: "` + f.Posts[0].Uuid.String() + `" photoID: "` + f.Files[1].UUID.String() + `"` +
		` 
			description: "new description"
			destination: {description:"dest" country:"dc" latitude:1.1 longitude:2.2}
			origin: {description:"origin" country:"oc" latitude:3.3 longitude:4.4}
			size: TINY
			neededAfter: "2019-11-01"
			neededBefore: "2019-12-25"
			category: "cat"
			url: "example.com" 
			cost: "1.00"
		`
	query := `mutation { post: updatePost(input: {` + input + `}) { id photo { id } description 
			destination { description country latitude longitude} 
			origin { description country latitude longitude}
			size neededAfter neededBefore category url cost isEditable}}`

	as.NoError(as.testGqlQuery(query, f.Users[0].Nickname, &postsResp))

	if err := as.DB.Load(&(f.Posts[0]), "PhotoFile", "Files"); err != nil {
		t.Errorf("failed to load post fixture, %s", err)
		t.FailNow()
	}

	as.Equal(f.Posts[0].Uuid.String(), postsResp.Post.ID)
	as.Equal(f.Files[1].UUID.String(), postsResp.Post.Photo.ID)
	as.Equal("new description", postsResp.Post.Description)
	as.Equal("dest", postsResp.Post.Destination.Description)
	as.Equal("dc", postsResp.Post.Destination.Country)
	as.Equal(1.1, postsResp.Post.Destination.Lat)
	as.Equal(2.2, postsResp.Post.Destination.Long)
	//as.Equal("origin", postsResp.Post.Origin.Description)
	//as.Equal("oc", postsResp.Post.Origin.Country)
	//as.Equal(3.3, postsResp.Post.Origin.Lat)
	//as.Equal(4.4, postsResp.Post.Origin.Long)
	as.Equal(models.PostSizeTiny, postsResp.Post.Size)
	as.Equal("2019-11-01T00:00:00Z", postsResp.Post.NeededAfter)
	as.Equal("2019-12-25T00:00:00Z", postsResp.Post.NeededBefore)
	as.Equal("cat", postsResp.Post.Category)
	as.Equal("example.com", postsResp.Post.Url)
	as.Equal("1", postsResp.Post.Cost)
	as.Equal(true, postsResp.Post.IsEditable)

	// Attempt to edit a locked post
	input = `id: "` + f.Posts[0].Uuid.String() + `" category: "new category"`
	query = `mutation { post: updatePost(input: {` + input + `}) { id status}}`

	as.Error(as.testGqlQuery(query, f.Users[1].Nickname, &postsResp))
}
