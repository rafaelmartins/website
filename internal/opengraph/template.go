package opengraph

type TemplateContext struct {
	Title       string
	Description string
	Image       string
}

func (og *OpenGraph) GetTemplateContext() TemplateContext {
	image := ""
	if og.pregenerated || (og.ogi != nil && og.ogi.template != "" && og.generate) || og.image != "" {
		image = url(og.baseurl)
	}
	return TemplateContext{
		Title:       og.title,
		Description: og.description,
		Image:       image,
	}
}
