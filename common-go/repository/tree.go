package repository

// SlugTree represents a tree structure for repository slugs and their data.
// The tree structure is like the following JSON format:
//     {
//       <saas>: {
//         <owner>: {
//           <repo-name>: <repo-data>,
//           ...
//         },
//         ...
//       },
//       ...
//     }
type SlugTree map[string]map[string]map[string]interface{}

// GetData returns a data for a specified repository if it exists.
// It returns nil otherwise.
func (tree SlugTree) GetData(slug *Slug) interface{} {
	if slug == nil {
		return nil
	}

	owners, ok := tree[slug.Saas]
	if !ok || owners == nil {
		return nil
	}

	repos, ok := owners[slug.Owner]
	if !ok || repos == nil {
		return nil
	}

	data, ok := repos[slug.Name]
	if !ok {
		return nil
	}

	return data
}

// PutData stores a data associated with a specified repository.
func (tree SlugTree) PutData(slug *Slug, data interface{}) {
	if slug == nil {
		return
	}

	owners, ok := tree[slug.Saas]
	if !ok || owners == nil {
		owners = make(map[string]map[string]interface{})
		tree[slug.Saas] = owners
	}

	repos, ok := owners[slug.Owner]
	if !ok || repos == nil {
		repos = make(map[string]interface{})
		owners[slug.Owner] = repos
	}

	repos[slug.Name] = data
}

// PutDataMulti stores a data associated with specified repositories.
func (tree SlugTree) PutDataMulti(slugs []*Slug, data interface{}) {
	for _, slug := range slugs {
		tree.PutData(slug, data)
	}
}

// Empty returns true if there is no data in the tree.
// It returns false otherwise.
func (tree SlugTree) Empty() bool {
	return len(tree) == 0
}

// ReferenceTree represents a tree structure for references and their data.
// The tree structure is like the following JSON format:
//     {
//       <ref-type>: {
//         <ref-name>: <ref-data>,
//         ...
//       },
//       ...
//     }
type ReferenceTree map[string]map[string]interface{}

// GetData returns a data for a specified reference if it exists.
// It returns nil otherwise.
func (tree ReferenceTree) GetData(ref *Reference) interface{} {
	if ref == nil {
		return nil
	}

	refs, ok := tree[ref.Type]
	if !ok || refs == nil {
		return nil
	}

	data, ok := refs[ref.Name]
	if !ok {
		return nil
	}

	return data
}

// PutData stores a data associated with a specified reference.
func (tree ReferenceTree) PutData(ref *Reference, data interface{}) {
	if ref == nil {
		return
	}

	refs, ok := tree[ref.Type]
	if !ok || refs == nil {
		refs = make(map[string]interface{})
		tree[ref.Type] = refs
	}

	refs[ref.Name] = data
}

// PutDataMulti stores a data associated with specified references.
func (tree ReferenceTree) PutDataMulti(refs []*Reference, data interface{}) {
	for _, ref := range refs {
		tree.PutData(ref, data)
	}
}

// Empty returns true if there is no data in the tree.
// It returns false otherwise.
func (tree ReferenceTree) Empty() bool {
	return len(tree) == 0
}
