/* eslint-disable react/display-name */
/* eslint-disable react/no-children-prop */
/* eslint-disable react/prop-types */
import React from 'react'
import './index.scss'
import titleNavigation from 'Util/markdownNavigation'
import MarkdownAccordion from '../../components/molecules/markdownAccordion'
import { MDXProvider } from '@mdx-js/react'
import Docs from 'Util/docs/index.md'

export default function DocsRoute () {
	return (
		<div className='docs-container'>
			<div className='banner'>
				<h1>Docs</h1>
			</div>
			<div className='markdown-container'>
				<div className='navigation'>
					<div className='sticky'>
						{titleNavigation.map(elm => {
							return (
								<MarkdownAccordion
									key={elm.id}
									title={elm.title}
									number={elm.number}
									id={elm.id}
									subtitle={elm.subtitle}
								/>
							)
						})}
					</div>
				</div>
				<div className='content'>
					<MDXProvider>
						<Docs />
					</MDXProvider>
				</div>
			</div>
		</div>
	)
}
