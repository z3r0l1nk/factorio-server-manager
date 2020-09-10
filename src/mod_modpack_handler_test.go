package main

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func SetupModPacks(t *testing.T, empty bool, emptyMods bool) {
	var err error

	// check if dev directory exists and create it
	if _, err = os.Stat("dev_packs"); os.IsNotExist(err) {
		err = os.Mkdir("dev_packs", 0775)
	}
	assert.NoError(t, err, `Error creating "dev_packs" directory`)

	if !empty {
		// check if dev directory exists and create it
		if _, err = os.Stat("dev_packs/test"); os.IsNotExist(err) {
			err = os.Mkdir("dev_packs/test", 0775)
		}
		assert.NoError(t, err, `Error creating "dev_packs/test" directory`)

		mods, err := newMods("dev_packs/test")
		assert.NoError(t, err, "error creating mods")

		if !emptyMods {
			err = mods.downloadMod("/download/belt-balancer/5e9f9db4bf9d30000c5303f2", "belt-balancer_2.1.3.zip", "belt-balancer")
			assert.NoError(t, err, `Error downloading Mod "belt-balancer"`)

			err = mods.downloadMod("/download/train-station-overview/5e8a0a8ee8864f000d0cb022", "train-station-overview_2.0.3.zip", "train-station-overview")
			assert.NoError(t, err, `Error downloading Mod "train-station-overview"`)
		}
	}
}

func CleanupModPacks(t *testing.T) {
	err := os.RemoveAll("dev_packs")
	assert.NoError(t, err, `Error removing directory "dev_packs"`)
}

func UnknownModpackTest(t *testing.T, method string, baseRoute string, route string, handlerFunc http.HandlerFunc) {
	t.Run("unknown modpack", func(t *testing.T) {
		SetupModPacks(t, true, true)
		defer CleanupModPacks(t)

		requestBody := strings.NewReader(`{"name": "belt-balancer"}`)

		CallRoute(t, method, baseRoute, route, requestBody, handlerFunc, http.StatusNotFound, "")
	})
}

func ModPackUnknownModTest(t *testing.T, method string, baseRoute string, route string, handlerFunc http.HandlerFunc) {
	t.Run("unknown mod", func(t *testing.T) {
		SetupModPacks(t, false, false)
		defer CleanupModPacks(t)

		requestBody := strings.NewReader(`{"name": "askhdbali"}`)

		CallRoute(t, method, baseRoute, route, requestBody, handlerFunc, http.StatusInternalServerError, "")
	})
}

func ModPackEmptyBodyTest(t *testing.T, method string, baseRoute string, route string, handlerFunc http.HandlerFunc) {
	t.Run("empty body", func(t *testing.T) {
		SetupModPacks(t, false, true)
		defer CleanupModPacks(t)
		SetupMods(t, false)
		defer CleanupMods(t)

		CallRoute(t, method, baseRoute, route, nil, handlerFunc, http.StatusBadRequest, "")
	})
}

func ModPackInvalidJsonBodyTest(t *testing.T, method string, baseRoute string, route string, handlerFunc http.HandlerFunc) {
	t.Run("invalid json body", func(t *testing.T) {
		SetupModPacks(t, false, true)
		defer CleanupModPacks(t)
		SetupMods(t, true)
		defer CleanupMods(t)

		requestBody := strings.NewReader(`{`)

		CallRoute(t, method, baseRoute, route, requestBody, handlerFunc, http.StatusBadRequest, "")
	})
}

func TestModPackListHandler(t *testing.T) {
	CheckShort(t)

	method := "GET"
	route := "/mods/packs/list"
	handlerFunc := ModPackListHandler

	t.Run("success", func(t *testing.T) {
		SetupModPacks(t, false, false)
		defer CleanupModPacks(t)

		expected := `{"mod_packs":[{"name":"test","mods":{"mods":[{"name":"belt-balancer","version":"2.1.3","title":"Belt Balancer","author":"knoxfighter","file_name":"belt-balancer_2.1.3.zip","factorio_version":"0.18.0.0","dependencies":null,"compatibility":true,"enabled":true},{"name":"train-station-overview","version":"2.0.3","title":"Train Station Overview","author":"knoxfighter","file_name":"train-station-overview_2.0.3.zip","factorio_version":"0.18.0.0","dependencies":null,"compatibility":true,"enabled":true}]}}]}`

		CallRoute(t, method, route, route, nil, handlerFunc, http.StatusOK, expected)
	})

	t.Run("empty modpack", func(t *testing.T) {
		SetupModPacks(t, true, false)
		defer CleanupModPacks(t)

		expected := `{"mod_packs": null}`

		CallRoute(t, method, route, route, nil, handlerFunc, http.StatusOK, expected)
	})
}

func TestModPackCreateHandler(t *testing.T) {
	CheckShort(t)

	method := "POST"
	route := "/mods/packs/create"
	handlerFunc := ModPackCreateHandler

	t.Run("success", func(t *testing.T) {
		SetupModPacks(t, true, false)
		defer CleanupModPacks(t)
		SetupMods(t, false)
		defer CleanupMods(t)

		requestBody := strings.NewReader(`{"name": "test"}`)
		expected := `{"mod_packs":[{"name":"test","mods":{"mods":[{"name":"belt-balancer","version":"2.1.3","title":"Belt Balancer","author":"knoxfighter","file_name":"belt-balancer_2.1.3.zip","factorio_version":"0.18.0.0","dependencies":null,"compatibility":true,"enabled":true},{"name":"train-station-overview","version":"2.0.3","title":"Train Station Overview","author":"knoxfighter","file_name":"train-station-overview_2.0.3.zip","factorio_version":"0.18.0.0","dependencies":null,"compatibility":true,"enabled":true}]}}]}`

		CallRoute(t, method, route, route, requestBody, handlerFunc, http.StatusOK, expected)
	})

	ModPackEmptyBodyTest(t, method, route, route, handlerFunc)

	ModPackInvalidJsonBodyTest(t, method, route, route, handlerFunc)
}

func TestModPackDeleteHandler(t *testing.T) {
	CheckShort(t)

	method := "POST"
	baseRoute := "/mods/packs/{modpack}/delete"
	route := "/mods/packs/test/delete"
	handlerFunc := ModPackDeleteHandler

	t.Run("success", func(t *testing.T) {
		SetupModPacks(t, false, false)
		defer CleanupModPacks(t)

		CallRoute(t, method, baseRoute, route, nil, handlerFunc, http.StatusOK, `"test"`)
	})

	UnknownModpackTest(t, method, baseRoute, route, handlerFunc)
}

func TestModPackLoadHandler(t *testing.T) {
	CheckShort(t)

	method := "POST"
	baseRoute := "/mods/packs/{modpack}/load"
	route := "/mods/packs/test/load"
	handlerFunc := ModPackLoadHandler

	t.Run("load mods", func(t *testing.T) {
		SetupModPacks(t, false, false)
		defer CleanupModPacks(t)
		SetupMods(t, true)
		defer CleanupMods(t)

		expected := `{"mods":[{"name":"belt-balancer","version":"2.1.3","title":"Belt Balancer","author":"knoxfighter","file_name":"belt-balancer_2.1.3.zip","factorio_version":"0.18.0.0","dependencies":null,"compatibility":true,"enabled":true},{"name":"train-station-overview","version":"2.0.3","title":"Train Station Overview","author":"knoxfighter","file_name":"train-station-overview_2.0.3.zip","factorio_version":"0.18.0.0","dependencies":null,"compatibility":true,"enabled":true}]}`

		CallRoute(t, method, baseRoute, route, nil, handlerFunc, http.StatusOK, expected)

		// check if mods are really loaded
		packMap, err := newModPackMap()
		assert.NoError(t, err, "Error creating modPackMap")

		mods, err := newMods("dev")
		assert.NoError(t, err, "Error creating mods object")

		packModsJson, err := json.Marshal(packMap["test"].Mods)
		assert.NoError(t, err, "Error marshalling mods from modPack")

		modsJson, err := json.Marshal(mods)
		assert.NoError(t, err, "Error marshalling mods object")

		assert.JSONEq(t, string(packModsJson), string(modsJson), "loaded mods and modPack are not identical")
	})

	t.Run("load empty modpack", func(t *testing.T) {
		SetupModPacks(t, false, true)
		defer CleanupModPacks(t)
		SetupMods(t, false)
		defer CleanupMods(t)

		expected := `{"mods":null}`

		CallRoute(t, method, baseRoute, route, nil, handlerFunc, http.StatusOK, expected)

		// check if mods are really loaded
		packMap, err := newModPackMap()
		assert.NoError(t, err, "Error creating modPackMap")

		mods, err := newMods("dev")
		assert.NoError(t, err, "Error creating mods object")

		packModsJson, err := json.Marshal(packMap["test"].Mods)
		assert.NoError(t, err, "Error marshalling mods from modPack")

		modsJson, err := json.Marshal(mods)
		assert.NoError(t, err, "Error marshalling mods object")

		assert.JSONEq(t, string(packModsJson), string(modsJson), "loaded mods and modPack are not identical")
	})

	UnknownModpackTest(t, method, baseRoute, route, handlerFunc)
}

func TestModPackModListHandler(t *testing.T) {
	CheckShort(t)

	method := "GET"
	baseRoute := "/mods/packs/{modpack}/list"
	route := "/mods/packs/test/list"
	handlerFunc := ModPackModListHandler

	t.Run("success", func(t *testing.T) {
		SetupModPacks(t, false, false)
		defer CleanupModPacks(t)

		expected := `{"mods":[{"name":"belt-balancer","version":"2.1.3","title":"Belt Balancer","author":"knoxfighter","file_name":"belt-balancer_2.1.3.zip","factorio_version":"0.18.0.0","dependencies":null,"compatibility":true,"enabled":true},{"name":"train-station-overview","version":"2.0.3","title":"Train Station Overview","author":"knoxfighter","file_name":"train-station-overview_2.0.3.zip","factorio_version":"0.18.0.0","dependencies":null,"compatibility":true,"enabled":true}]}`

		CallRoute(t, method, baseRoute, route, nil, handlerFunc, http.StatusOK, expected)
	})

	UnknownModpackTest(t, method, baseRoute, route, handlerFunc)
}

func TestModPackModToggleHandler(t *testing.T) {
	CheckShort(t)

	method := "POST"
	baseRoute := "/mods/packs/{modpack}/mod/toggle"
	route := "/mods/packs/test/mod/toggle"
	handlerFunc := ModPackModToggleHandler

	t.Run("success", func(t *testing.T) {
		SetupModPacks(t, false, false)
		defer CleanupModPacks(t)

		requestBody := strings.NewReader(`{"name": "belt-balancer"}`)

		// mod is now deactivated
		expected := "false"

		CallRoute(t, method, baseRoute, route, requestBody, handlerFunc, http.StatusOK, expected)

		// check if changes happened
		packMap, err := newModPackMap()
		assert.NoError(t, err, "Error creating modPackMap")

		found := false
		for _, mod := range packMap["test"].Mods.ModSimpleList.Mods {
			if mod.Name == "belt-balancer" {
				// this mod has to be deactivated now
				if mod.Enabled {
					t.Fatalf("Mod is wrongly enabled, it should be disabled by now")
				}
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Mod not found")
		}

		// toggle again, to check if the other direction also works
		// mod is now activated again
		expected = "true"

		// reset request body, it has to be red again
		requestBody.Seek(0, 0)

		CallRoute(t, method, baseRoute, route, requestBody, handlerFunc, http.StatusOK, expected)

		packMap, err = newModPackMap()
		assert.NoError(t, err, "Error creating modPackMap")

		found = false
		for _, mod := range packMap["test"].Mods.ModSimpleList.Mods {
			if mod.Name == "belt-balancer" {
				// this mod has to be deactivated now
				if !mod.Enabled {
					t.Fatalf("Mod is wrongly disabled, it should be enabled again")
				}
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Mod not found")
		}
	})

	UnknownModpackTest(t, method, baseRoute, route, handlerFunc)

	ModPackUnknownModTest(t, method, baseRoute, route, handlerFunc)

	ModPackEmptyBodyTest(t, method, baseRoute, route, handlerFunc)

	ModPackInvalidJsonBodyTest(t, method, baseRoute, route, handlerFunc)
}

func TestModPackModDeleteHandler(t *testing.T) {
	CheckShort(t)

	method := "POST"
	baseRoute := "/mods/packs/{modpack}/mod/delete"
	route := "/mods/packs/test/mod/delete"
	handlerFunc := ModPackModDeleteHandler

	t.Run("success", func(t *testing.T) {
		SetupModPacks(t, false, false)
		defer CleanupModPacks(t)

		requestBody := strings.NewReader(`{"name": "belt-balancer"}`)

		CallRoute(t, method, baseRoute, route, requestBody, handlerFunc, http.StatusOK, `true`)

		// check if mod is really not installed anymore
		packMap, err := newModPackMap()
		assert.NoError(t, err, "Error creating modPackMap")

		if packMap["test"].Mods.ModSimpleList.checkModExists("belt-balancer") {
			t.Fatalf("Mod is still installed, it should be gone by now")
		}
	})

	UnknownModpackTest(t, method, baseRoute, route, handlerFunc)

	ModPackUnknownModTest(t, method, baseRoute, route, handlerFunc)

	ModPackEmptyBodyTest(t, method, baseRoute, route, handlerFunc)

	ModPackInvalidJsonBodyTest(t, method, baseRoute, route, handlerFunc)
}

func TestModPackModDeleteAllHandler(t *testing.T) {
	CheckShort(t)

	method := "POST"
	baseRoute := "/mods/packs/{modpack}/mod/delete/all"
	route := "/mods/packs/test/mod/delete/all"
	handlerFunc := ModPackModDeleteAllHandler

	t.Run("success", func(t *testing.T) {
		SetupModPacks(t, false, false)
		defer CleanupModPacks(t)

		CallRoute(t, method, baseRoute, route, nil, handlerFunc, http.StatusOK, "true")

		// check if really empty
		packMap, err := newModPackMap()
		assert.NoError(t, err, "Error creating modPackMap")

		if len(packMap["test"].Mods.ModInfoList.Mods) != 0 {
			t.Fatal("There are still mods in the modpack")
		}
	})

	UnknownModpackTest(t, method, baseRoute, route, handlerFunc)
}

func TestModPackModUpdateHandler(t *testing.T) {
	CheckShort(t)

	method := "POST"
	baseRoute := "/mods/packs/{modpack}/mod/update"
	route := "/mods/packs/test/mod/update"
	handlerFunc := ModPackModUpdateHandler

	requestBodySuccess := `{"modName": "belt-balancer", "downloadUrl": "/download/belt-balancer/5e711cd95bcf4f000b96b22c", "fileName": "belt-balancer_2.1.2.zip"}`

	t.Run("success", func(t *testing.T) {
		SetupModPacks(t, false, false)
		defer CleanupModPacks(t)

		expected := `{"name":"belt-balancer","version":"2.1.2","title":"Belt Balancer","author":"knoxfighter","file_name":"belt-balancer_2.1.2.zip","factorio_version":"0.18.0.0","dependencies":null,"compatibility":true,"enabled":true}`

		CallRoute(t, method, baseRoute, route, strings.NewReader(requestBodySuccess), handlerFunc, http.StatusOK, expected)
	})

	t.Run("success with disabled mod", func(t *testing.T) {
		SetupModPacks(t, false, false)
		defer CleanupModPacks(t)

		// disable "belt-balancer" mod, so we can test, if it is still deactivated after
		packMap, err := newModPackMap()
		assert.NoError(t, err, "Error creating modPackMap")

		err, _ = packMap["test"].Mods.ModSimpleList.toggleMod("belt-balancer")
		assert.NoError(t, err, "Error toggling mod")

		expected := `{"name":"belt-balancer","version":"2.1.2","title":"Belt Balancer","author":"knoxfighter","file_name":"belt-balancer_2.1.2.zip","factorio_version":"0.18.0.0","dependencies":null,"compatibility":true,"enabled":false}`

		CallRoute(t, method, baseRoute, route, strings.NewReader(requestBodySuccess), handlerFunc, http.StatusOK, expected)
	})

	UnknownModpackTest(t, method, baseRoute, route, handlerFunc)

	ModPackEmptyBodyTest(t, method, baseRoute, route, handlerFunc)

	ModPackInvalidJsonBodyTest(t, method, baseRoute, route, handlerFunc)

	t.Run("unknown mod", func(t *testing.T) {
		SetupModPacks(t, false, false)
		defer CleanupModPacks(t)

		requestBody := `{"modName": "asldbsac", "downloadUrl": "/download/belt-balancer/5e711cd95bcf4f000b96b22c", "fileName": "belt-balancer_2.1.2.zip"}`

		CallRoute(t, method, baseRoute, route, strings.NewReader(requestBody), handlerFunc, http.StatusNotFound, "")
	})

	t.Run("wrong download link", func(t *testing.T) {
		SetupModPacks(t, false, false)
		defer CleanupModPacks(t)

		requestBody := `{"modName": "asldbsac", "downloadUrl": "/download/belt-balancer/95bcf4f000b96b22c", "fileName": "belt-balancer_2.1.2.zip"}`

		CallRoute(t, method, baseRoute, route, strings.NewReader(requestBody), handlerFunc, http.StatusInternalServerError, "")

		// check if old mod is still there
		packMap, err := newModPackMap()
		assert.NoError(t, err, "Error creating modPackMap")

		var found = false
		for _, mod := range packMap["test"].Mods.ModInfoList.Mods {
			if mod.Name == "belt-balancer" {
				found = true
			}
		}

		if !found {
			t.Fatal(`Mod "belt-balancer" is not there anymore`)
		}
	})
}

func ModPackModUploadRequest(t *testing.T, body bool, filePath string) *httptest.ResponseRecorder {
	method := "POST"
	baseRoute := "/mods/packs/{modpack}/mod/upload"
	route := "/mods/packs/test/mod/upload"
	handlerFunc := ModPackModUploadHandler

	var err error

	requestBody := &bytes.Buffer{}

	writer := multipart.NewWriter(requestBody)

	if body {
		file, err := os.Open(filePath)
		if err == nil {
			assert.NoError(t, err, "error opening mod file")

			formFile, err := writer.CreateFormFile("mod_file", filepath.Base(filePath))
			assert.NoError(t, err, "error creating formFileWriter")

			_, err = io.Copy(formFile, file)
			assert.NoError(t, err, "error copying file to form")
		}
	}

	err = writer.Close()
	if err != nil {
		t.Fatalf("error closing the multipart writer: %s", err)
	}

	// create request to send
	request, err := http.NewRequest(method, route, requestBody)
	assert.NoError(t, err, "Error creating request")
	request.Header.Set("Content-Type", writer.FormDataContentType())

	// create response recorder
	recorder := httptest.NewRecorder()

	// get the handler, where the request is handled
	router := mux.NewRouter()
	router.HandleFunc(baseRoute, handlerFunc)

	// call the handler directly
	router.ServeHTTP(recorder, request)

	return recorder
}

func TestModPackModUploadHandler(t *testing.T) {
	CheckShort(t)

	method := "POST"
	baseRoute := "/mods/packs/{modpack}/mod/upload"
	route := "/mods/packs/test/mod/upload"
	handlerFunc := ModPackModUploadHandler

	t.Run("success", func(t *testing.T) {
		SetupModPacks(t, false, true)
		defer CleanupModPacks(t)

		recorder := ModPackModUploadRequest(t, true, "factorio_testfiles/belt-balancer_2.1.3.zip")

		// status has to be 200
		if recorder.Code != http.StatusOK {
			t.Fatalf("Wrong Status Code. expected %v - got %v", http.StatusOK, recorder.Code)
		}

		// check if mod is uploaded correctly
		packMap, err := newModPackMap()
		assert.NoError(t, err, "error creating modPackMap")

		expected := ModsResultList{
			ModsResult: []ModsResult{
				{
					ModInfo: ModInfo{
						Name:            "belt-balancer",
						Version:         "2.1.3",
						Title:           "Belt Balancer",
						Author:          "knoxfighter",
						FileName:        "belt-balancer_2.1.3.zip",
						FactorioVersion: Version{0, 18, 0, 0},
						Dependencies:    nil,
						Compatibility:   true,
					},
					Enabled: true,
				},
			},
		}

		actual := packMap["test"].Mods.listInstalledMods()
		assert.Equal(t, expected, actual, `New mod is not correctly installed. expected "%v" - actual "%v"`, expected, actual)
	})

	ModPackEmptyBodyTest(t, method, baseRoute, route, handlerFunc)

	t.Run("empty file", func(t *testing.T) {
		SetupModPacks(t, false, true)
		defer CleanupModPacks(t)

		recorder := ModPackModUploadRequest(t, true, "")
		assert.Equal(t, http.StatusBadRequest, recorder.Code, "wrong response code.")
	})

	t.Run("invalid mod file (txt-file)", func(t *testing.T) {
		SetupModPacks(t, false, true)
		defer CleanupModPacks(t)

		recorder := ModPackModUploadRequest(t, false, "factorio_testfiles/file_usage.txt")
		assert.Equal(t, http.StatusBadRequest, recorder.Code, "wrong response code.")
	})

	t.Run("invalid mod file (zip-file)", func(t *testing.T) {
		SetupModPacks(t, false, true)
		defer CleanupModPacks(t)

		recorder := ModPackModUploadRequest(t, true, "factorio_testfiles/invalid_mod.zip")
		assert.Equal(t, http.StatusInternalServerError, recorder.Code, "wrong response code.")
	})
}

func TestModPackModPortalInstallHandler(t *testing.T) {
	CheckShort(t)

	method := "POST"
	baseRoute := "/mods/packs/{modpack}/portal/install"
	route := "/mods/packs/test/portal/install"
	handlerFunc := ModPackModPortalInstallHandler

	t.Run("success", func(t *testing.T) {
		SetupModPacks(t, false, true)
		defer CleanupModPacks(t)

		requestBody := strings.NewReader(`{"modName": "belt-balancer", "downloadUrl": "/download/belt-balancer/5e711cd95bcf4f000b96b22c", "fileName": "belt-balancer_2.1.2.zip"}`)

		expected := `{"mods":[{"name":"belt-balancer","version":"2.1.2","title":"Belt Balancer","author":"knoxfighter","file_name":"belt-balancer_2.1.2.zip","factorio_version":"0.18.0.0","dependencies":null,"compatibility":true,"enabled":true}]}`

		CallRoute(t, method, baseRoute, route, requestBody, handlerFunc, http.StatusOK, expected)
	})

	ModPackEmptyBodyTest(t, method, baseRoute, route, handlerFunc)

	ModPackInvalidJsonBodyTest(t, method, baseRoute, route, handlerFunc)

	t.Run("wrong download link", func(t *testing.T) {
		SetupModPacks(t, false, true)
		defer CleanupModPacks(t)

		requestBody := strings.NewReader(`{"modName": "belt-balancer", "downloadUrl": "/download/belt-balancer/95bcf4f000b96b22c", "fileName": "belt-balancer_2.1.2.zip"}`)

		CallRoute(t, method, baseRoute, route, requestBody, handlerFunc, http.StatusInternalServerError, "")
	})
}

func TestModPackModPortalInstallMultipleHandler(t *testing.T) {
	CheckShort(t)

	method := "POST"
	baseRoute := "/mods/packs/{modpack}/portal/install/multiple"
	route := "/mods/packs/test/portal/install/multiple"
	handlerFunc := ModPackModPortalInstallMultipleHandler

	t.Run("success", func(t *testing.T) {
		SetupModPacks(t, false, true)
		defer CleanupModPacks(t)

		requestBody := strings.NewReader(`[{"name": "belt-balancer", "version": "2.1.2"}, {"name": "train-station-overview", "version": "2.0.2"}]`)

		expected := `{"mods":[{"name":"belt-balancer","version":"2.1.2","title":"Belt Balancer","author":"knoxfighter","file_name":"belt-balancer_2.1.2.zip","factorio_version":"0.18.0.0","dependencies":null,"compatibility":true,"enabled":true},{"name":"train-station-overview","version":"2.0.2","title":"Train Station Overview","author":"knoxfighter","file_name":"train-station-overview_2.0.2.zip","factorio_version":"0.18.0.0","dependencies":null,"compatibility":true,"enabled":true}]}`

		CallRoute(t, method, baseRoute, route, requestBody, handlerFunc, http.StatusOK, expected)
	})

	t.Run("unknown mod", func(t *testing.T) {
		SetupModPacks(t, false, true)
		defer CleanupModPacks(t)

		requestBody := strings.NewReader(`[{"name": "askdhcb", "version": "2.1.2"}]`)

		CallRoute(t, method, baseRoute, route, requestBody, handlerFunc, http.StatusInternalServerError, "")
	})

	t.Run("unknown version", func(t *testing.T) {
		SetupModPacks(t, false, true)
		defer CleanupModPacks(t)

		requestBody := strings.NewReader(`[{"name": "belt-balancer", "version": "0.1.12"}]`)

		CallRoute(t, method, baseRoute, route, requestBody, handlerFunc, http.StatusInternalServerError, "")
	})

	ModPackEmptyBodyTest(t, method, baseRoute, route, handlerFunc)

	ModPackInvalidJsonBodyTest(t, method, baseRoute, route, handlerFunc)
}
