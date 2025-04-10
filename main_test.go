package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Main", func() {
	Describe("normalizeDescription()", func() {
		DescribeTable("normalization of markdown",
			func(input, output string) {
				Expect(normalizeDescription(input)).To(Equal(output))
			},
			Entry("Blank input", "", ""),
			Entry("Plain text input", "Hello world", "Hello world"),
			Entry("Markdown input without comments", "**Hello world**\n\n* test\n*test2", "**Hello world**\n\n* test\n*test2"),
			Entry("Markdown with comments", "**Hello world**\n\n<!--- \nremove this if no breaking changes -->", "**Hello world**"),
			Entry("Markdown with headings", 
				`## Change Management
				<!-- Do not delete this section, explicitly use 'N/A' or 'None' instead.
					What does biz/ops need to know about this change? How will this impact customers or internal users? 
					Are there any migrations or infrastructure changes required? -->
					blah blah blah
				## Engineering Details
					<!-- Engineering details and discussions in full glory. -->
				## See Also
					<!-- Related Github links, documentation, logs, Slack comments, etc. -->
					<!-- Jira issue, e.g.: CRD-nnn -->`, 
				`## Change Management
					## Engineering Details
					## See Also`),
		)
	})
})
 