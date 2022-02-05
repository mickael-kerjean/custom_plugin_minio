package plg_search_elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
	. "github.com/mickael-kerjean/filestash/server/common"
	"os"
	"strings"
)

type ElasticSearch struct {
	Es7          *elasticsearch7.Client
	Index        string
	PathField    string
	ContentField string
	SizeField    string
}

func init() {
	if plugin_enable := Config.Get("features.elasticsearch.enable").Schema(func(f *FormElement) *FormElement {
		if f == nil {
			f = &FormElement{}
		}
		f.Default = false
		f.Name = "enable"
		f.Type = "boolean"
		f.Target = []string{}
		f.Description = "Enable/Disable search using ElasticSearch. Please note that indexing is not handled by this plugin. This setting requires a restart to come into effect."
		f.Placeholder = "Default: false"
		if u := os.Getenv("ELASTICSEARCH_URL"); u != "" {
			f.Default = true
		}
		return f
	}).Bool(); plugin_enable == false {
		return
	}
	Config.Get("features.elasticsearch.url").Schema(func(f *FormElement) *FormElement {
		if f == nil {
			f = &FormElement{}
		}
		f.Id = "url"
		f.Name = "url"
		f.Type = "text"
		f.Description = "Location of your ElasticSearch server(s)"
		f.Default = ""
		f.Placeholder = "Eg: http://127.0.0.1:9200[,http://127.0.0.1:9201]"
		if u := os.Getenv("ELASTICSEARCH_URL"); u != "" {
			f.Default = u
			f.Placeholder = fmt.Sprintf("Default: '%s'", u)
		}
		return f
	})
	Config.Get("features.elasticsearch.index").Schema(func(f *FormElement) *FormElement {
		if f == nil {
			f = &FormElement{}
		}
		f.Id = "index"
		f.Name = "index"
		f.Type = "text"
		f.Description = "Name of the Elasticsearch index"
		f.Default = ""
		f.Placeholder = "Eg: filestash_index"
		if u := os.Getenv("ELASTICSEARCH_INDEX"); u != "" {
			f.Default = u
			f.Placeholder = fmt.Sprintf("Default: '%s'", u)
		}
		return f
	})
	Config.Get("features.elasticsearch.username").Schema(func(f *FormElement) *FormElement {
		if f == nil {
			f = &FormElement{}
		}
		f.Id = "username"
		f.Name = "username"
		f.Type = "text"
		f.Description = "Username for connecting to Elasticsearch"
		f.Default = ""
		f.Placeholder = "Eg: username"
		if u := os.Getenv("ELASTICSEARCH_USERNAME"); u != "" {
			f.Default = u
			f.Placeholder = fmt.Sprintf("Default: '%s'", u)
		}
		return f
	})
	Config.Get("features.elasticsearch.password").Schema(func(f *FormElement) *FormElement {
		if f == nil {
			f = &FormElement{}
		}
		f.Id = "password"
		f.Name = "password"
		f.Type = "text"
		f.Description = "Password for connecting to Elasticsearch"
		f.Default = ""
		f.Placeholder = "Eg: password"
		if u := os.Getenv("ELASTICSEARCH_USERNAME"); u != "" {
			f.Default = u
			f.Placeholder = fmt.Sprintf("Default: '%s'", u)
		}
		return f
	})
	Config.Get("features.elasticsearch.field_path").Schema(func(f *FormElement) *FormElement {
		if f == nil {
			f = &FormElement{}
		}
		f.Id = "field_path"
		f.Name = "field_path"
		f.Type = "text"
		f.Description = "Field name for file path"
		f.Default = ""
		f.Placeholder = "Eg: path_field"
		if u := os.Getenv("ELASTICSEARCH_FIELD_PATH"); u != "" {
			f.Default = u
			f.Placeholder = fmt.Sprintf("Default: '%s'", u)
		}
		return f
	})
	Config.Get("features.elasticsearch.field_content").Schema(func(f *FormElement) *FormElement {
		if f == nil {
			f = &FormElement{}
		}
		f.Id = "field_content"
		f.Name = "field_content"
		f.Type = "text"
		f.Description = "Field name for file content"
		f.Default = ""
		f.Placeholder = "Eg: content_field"
		if u := os.Getenv("ELASTICSEARCH_FIELD_CONTENT"); u != "" {
			f.Default = u
			f.Placeholder = fmt.Sprintf("Default: '%s'", u)
		}
		return f
	})
	Config.Get("features.elasticsearch.field_size").Schema(func(f *FormElement) *FormElement {
		if f == nil {
			f = &FormElement{}
		}
		f.Id = "field_size"
		f.Name = "field_size"
		f.Type = "text"
		f.Description = "Field name for file size"
		f.Default = ""
		f.Placeholder = "Eg: size_field"
		if u := os.Getenv("ELASTICSEARCH_FIELD_SIZE"); u != "" {
			f.Default = u
			f.Placeholder = fmt.Sprintf("Default: '%s'", u)
		}
		return f
	})
	Config.Get("features.elasticsearch.enable_root_search").Schema(func(f *FormElement) *FormElement {
		if f == nil {
			f = &FormElement{}
		}
		f.Id = "enable_root_search"
		f.Name = "enable_root_search"
		f.Type = "boolean"
		f.Description = "Enable searching from root level (could potentially bypass access control restrictions)"
		f.Default = false
		return f
	})

	cfg := elasticsearch7.Config{
		Addresses: strings.Split(Config.Get("features.elasticsearch.url").String(), ","),
	}
	if Config.Get("features.elasticsearch.username").String() != "" {
		cfg.Username = Config.Get("features.elasticsearch.username").String()
	}
	if Config.Get("features.elasticsearch.password").String() != "" {
		cfg.Password = Config.Get("features.elasticsearch.password").String()
	}

	var (
		r map[string]interface{}
	)

	es7, err := elasticsearch7.NewDefaultClient()
	if err != nil {
		Log.Error("ES::init Error creating elasticsearch client: %s", err)
		return
	}
	res, err := es7.Info()
	if err != nil {
		Log.Error("ES::init Error getting response: %s", err)
		return
	}
	defer res.Body.Close()
	// Check response status
	if res.IsError() {
		Log.Error("ES::init Error: %s", res.String())
		return
	}
	// Deserialize the response into a map.
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		Log.Error("ES::init Error parsing the response body: %s", err)
		return
	}
	// Print client and server version numbers.
	Log.Debug("ES::init Client: %s", elasticsearch7.Version)
	Log.Debug("ES::init Server: %s", r["version"].(map[string]interface{})["number"])
	Log.Debug(strings.Repeat("~", 37))

	es := &ElasticSearch{
		Es7:          es7,
		Index:        Config.Get("features.elasticsearch.index").String(),
		PathField:    Config.Get("features.elasticsearch.field_path").String(),
		ContentField: Config.Get("features.elasticsearch.field_content").String(),
		SizeField:    Config.Get("features.elasticsearch.field_size").String(),
	}

	Hooks.Register.SearchEngine(es)
}

func (this ElasticSearch) Query(app App, path string, keyword string) ([]IFile, error) {
	Log.Debug("ES::Query path: %s, keyword, %s", path, keyword)
	if path == "/" {
		if Config.Get("features.elasticsearch.enable_root_search").Bool() {
			path = "*"
		} else {
			return nil, NewError("Cannot search from root level.", 404)
		}
	}
	var (
		r map[string]interface{}
	)

	// Build the request body.
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"query_string": map[string]interface{}{
				"fields": [2]string{this.ContentField, this.PathField},
				"query":  "(" + this.PathField + ":" + strings.ReplaceAll(path, "/", "\\/") + ") AND (" + keyword + ")",
			},
		},
		"highlight": map[string]interface{}{
			"fields": map[string]interface{}{
				"attachment.content": map[string]interface{}{},
			},
		},
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		Log.Error("ES::Query query_builder: Error encoding query: %s", err)
		return nil, ErrNotFound
	}

	// Perform the search request.
	res, err := this.Es7.Search(
		this.Es7.Search.WithContext(context.Background()),
		this.Es7.Search.WithIndex(this.Index),
		this.Es7.Search.WithBody(&buf),
		this.Es7.Search.WithTrackTotalHits(true),
		this.Es7.Search.WithPretty(),
	)

	if err != nil {
		Log.Error("ES::Query search: Error getting response: %s", err)
		res.Body.Close()
		return nil, ErrNotFound
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			Log.Error("ES::Query search: Error parsing the response body: %s", err)
			res.Body.Close()
			return nil, ErrNotFound
		} else {
			// Print the response status and error information.
			Log.Debug("ES::Query search: [%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
			return nil, NewError(e["error"].(map[string]interface{})["reason"].(string), 404)
		}
	}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		Log.Error("ES::Query search: Error parsing the response body: %s", err)
		res.Body.Close()
		return nil, ErrNotFound
	}
	// Print the response status, number of results, and request duration.
	Log.Debug(
		"ES::Query search: [%s] %d hits; took: %dms",
		res.Status(),
		int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
		int(r["took"].(float64)),
	)

	files := []IFile{}

	// Print the ID and document source for each hit.
	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {

		Log.Debug("ES::Query search: * highlights %v", hit.(map[string]interface{})["highlight"])

		size := hit.(map[string]interface{})["_source"].(map[string]interface{})[this.SizeField].(float64)

		resPath := hit.(map[string]interface{})["_source"].(map[string]interface{})[this.PathField].(string)
		pathTokens := strings.Split(resPath, "/")
		resFilename := pathTokens[len(pathTokens)-1]
		fileToken := strings.Split(resFilename, ".")
		resExt := ""
		if len(fileToken) > 1 {
			resExt = fileToken[len(fileToken)-1]
		}

		Log.Debug("ES::Query search: * ID=%s, path=%s, FName=%s, ext=%s",
			hit.(map[string]interface{})["_id"],
			resPath,
			resFilename,
			resExt)

		files = append(files, File{
			FName: resFilename,
			FType: "file", // ENUM("file", "directory")
			FSize: int64(size),
			FPath: resPath,
		})
	}
	Log.Debug(strings.Repeat("=", 37))

	return files, nil
}
