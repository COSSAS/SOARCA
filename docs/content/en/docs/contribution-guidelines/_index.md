---
title: Contribution Guidelines
weight: 7
description: How to contribute to SOARCA
---

SOARCA is an open-source project written in [Golang](https://go.dev/) and we love getting patches and contributions, and feature suggestions to make SOARCA and its docs even better. We welcome participation from anyone, regardless of their affiliation with OASIS. We invite constructive contributions and feedback from all contributors, following the [standard practices](https://docs.github.com/en/get-started/exploring-projects-on-github/contributing-to-a-project) for participation in GitHub public repository projects.

We expect everyone to follow our [Code of Conduct](/docs/contribution-guidelines/code_of_conduct/), the licenses for each repository, and agree to our Contributor License Agreement when you make your first contribution.

Thank you for contributing to our project! Your efforts make a difference.

## Contributing to SOARCA

The SOARCA itself lives on [github](https://github.com/COSSAS/SOARCA).

## How to contribute

Before making contributions to the project repositories, please follow these general steps for [GitHub contribution](https://docs.github.com/en/get-started/exploring-projects-on-github/contributing-to-a-project). 

### I found a bug / Creating issues

If there's something you'd like to see in SOARCA (or if you've found something that isn't working the way you'd expect), but you're not sure how to fix it yourself, please create an [issue](https://github.com/COSSAS/SOARCA/issues/new/choose). Make sure to adhere to the structure of an issue submission. Fully comprehend the problem at hand and provide comprehensive details in your issue description.


{{% alert title="Security issues" color="warning" %}}
For security issues, we kindly request that you refrain from reporting them using the issue tracker. Instead, please contact us directly: [slack](https://join.slack.com/t/cossas/shared_invite/zt-2i4zxg0oh-dhhL4zTSX5olysngrPxDkg)
{{% /alert %}}


### Feature additions or requests

You can submit feature requests either through [GitHub issues](https://github.com/COSSAS/SOARCA/issues) or the [discussion pages](https://github.com/COSSAS/SOARCA/discussions).

### Code reviews

Every submission, including those from project members, must undergo review and approval from at least one core maintainer. GitHub pull requests are utilized for this process. Consult [GitHub Help](https://help.github.com/articles/about-pull-requests/) for more
information on using pull requests.

### Branch naming

The CI is configured to only allow for certain branch naming namely:
- master
- development
- feature/<your feature name here>
- feature/docs/<your feature to update docs>
- bugfix/<your bugfix here>
- release/x.x
- hotfix/<your hotfix on a release branch>

### Coding style

The project has opted to select the [go style guide](https://google.github.io/styleguide/go/) with some exceptions:
- Receiver name are not one letter https://google.github.io/styleguide/go/decisions#receiver-names so use `info` instead of `i` 
- Initialisms are CamelCase https://google.github.io/styleguide/go/decisions#initialisms so use `Xml` instead of `XML`

## Communication channels

Feel free to engage with the community for discussions and assistance via one of the following channels:

- [slack](https://join.slack.com/t/cossas/shared_invite/zt-2i4zxg0oh-dhhL4zTSX5olysngrPxDkg)
- [GitHub discussions](https://github.com/COSSAS/SOARCA/discussions)

## Contributing to these docs
 
Would you like to enhance our documentation? Our documentation is built using the [Hugo framework](https://gohugo.io/) along with the [Docsy theme](https://github.com/google/docsy) template.

### Quick start with Hugo and Docsy

1. Install Hugo; the installation guide can be found [here](https://gohugo.io/getting-started/quick-start/).
2. Clone our repository, and if you make changes, fork our repository. Use the following command to clone: `git clone <repository_url>`.
3. All the documentation for the GitHub Pages lives under `/documentation`. Use the `cd documentation && hugo serve` command to preview the documentation locally. Open `http://localhost:1313` in your web browser to view the documentation. In most cases, docsy will automatically reload the site to reflect any changes to the documentation or the code. Changes to some parts of the docsy code may require manually reloading the page or restarting the container.
4. Continue with the usual GitHub workflow to edit files, commit them, push the changes up to your fork, and create a pull request.


#### Updating a single page

If you've just spotted something you'd like to change while using the docs, Docsy has a shortcut for you:

1. Click **Edit this page** in the top right-hand corner of the page.
1. If you don't already have an up-to-date fork of the project repo, you are prompted to get one - click **Fork this repository and propose changes** or **Update your Fork** to get an up-to-date version of the project to edit. The appropriate page in your fork is displayed in edit mode.


## License 

The project is licensed under the Apache License 2.0. See full license [here](https://github.com/COSSAS/SOARCA/blob/development/LICENSE).