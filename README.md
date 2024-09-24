# website
Yet another NIH static website framework.

The idea is to have a single website for **ALL** my stuff: blog, projects, talks, random notes, etc., and this simple framework is probably flexible enough for that.

*This is a work in progress.*

## Why?
For a long time I've been willing to have simple project pages for my open source projects, integrated to my website/blog. The easiest way of doing that would be to write proper `README`s in the Git repositories and have them integrated to my website, this way both the repositories and my website would be proper documentation points, with all the data synchronized. However, implementing such advanced features that require API interaction is not trivial in [blogc](https://blogc.rgm.io/). Or maybe I'm getting too old to implement parsers/renderers in C… But in the end it was easier/faster to just implemented what I wanted in Go over the past months. Also, the standard Go templates library provides flexibility enough to handle a huge variety of page formats from the same template base, which is something I will need.

The code is kinda generic (writing code that way is just stronger than me...), but that’s it: there’s no documentation neither usage examples, and my content repository is private. This program is open source, but if you decide to use it, you are on your own.

## License
This code is released under a [BSD 3-Clause License](LICENSE).
