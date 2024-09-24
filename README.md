# website
Yet another NIH static website framework.

The idea is to have a single website for **ALL** my stuff: blog, projects, talks, random notes, etc..

*This is a work in progress.*

## Why?
For a long time I've been willing to have simple project pages for my open source projects, integrated to my website/blog. The easiest way of doing that would be to write proper `README.md`s in the Git repositories and have them rendered and included to my website. This way both my website and the repositories would be proper documentation endpoints, with all the data synchronized. However, implementing such advanced features that require API interaction is not trivial in [blogc](https://blogc.rgm.io/). Or maybe I'm getting too old to implement parsers/renderers in C, who knows… But in the end it was easier to just slowly implement what I wanted in Go over the past few months. Also, the standard Go `html/template` library allows to easily handle a variety of post/page formats from the same base templates, which is something I suppose that I'll need.

The code is kinda generic (writing code that way is just stronger than me...), but that’s it: there’s no documentation neither usage examples, and my content repository is private. This program is open source, but if you decide to use it, you are on your own. There are some quite interesting code snippets in this codebae, though. Make sure to take a look if you like Go `:-)`.

## License
This code is released under a [BSD 3-Clause License](LICENSE).
