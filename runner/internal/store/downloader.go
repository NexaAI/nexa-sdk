package store

const HF_ENDPOINT = "https://huggingface.co"

//func GetModelFiles(repoId string) (*types.Model, error) {
//	req, e := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s/tree/main", HF_ENDPOINT, repoId), nil)
//	if e != nil {
//		return nil, e
//	}
//
//	r, e := http.DefaultClient.Do(req)
//	if e != nil {
//		return nil, e
//	}
//	fmt.Println(r)
//
//	return nil, nil
//
//}

// TODO: multi gguf file
func (s *Store) Pull(repoId string) {

}
