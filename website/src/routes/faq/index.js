import React, { useEffect, useState } from 'react'
import Pagination from 'Components/atoms/pagination'
import Accordion from 'Components/molecules/accordion'
import Search from 'Components/atoms/search'
import { search } from 'fast-fuzzy'

import './index.scss'

const faqArray = [
	{
		question: 'Q: What is the difference between UnDistro, EKS, AKS and GKE?',
		answer: `A: Think of UnDistro as an additional “layer” to help you with day 2 operations, 
		no matter the infrastructure or if it is a hosted or self-hosted Kubernetes. Choosing to 
		use unDistro is not a matter of choosing unDistro over any hosted Kubernetes offer. 
		UnDistro is a software that runs inside your Kubernetes management cluster.`
	},

	{
		question: 'Q: I already use a managed kubernetes. Why should I use UnDistro?',
		answer: `A: UnDistro uses the kubernetes reconciliation process to ensure that 
		your cluster is always in the desired state with respect to kubernetes version, 
		number of nodes, size of nodes and firewall policies for the cluster with the 
		infrastructure provider.`
	},
	{
		question: 'Q: Does the Kubernetes API need to be publicly available?',
		answer: 'A: No, but the API needs to be accessible from the management cluster.'
	},
	{
		question: 'Q: Can an UnDistro upgrade compromise my clusters?',
		answer: 'A: No, the upgrade operations take place in the management cluster, your workload clusters continue to work with no impact.'
	},
	{
		question: 'Q: Does the loss of communication between the UnDistro management cluster and my cluster cause any problems?',
		answer: 'A: You can expect some issues if you need to scale your cluster, otherwise the clusters will continue to work properly.'
	},
	{
		question: 'Q: Can I install UnDistro on any Linux distribution?',
		answer: 'A: UnDistro CLI runs on any Linux and Darwin OS and its components are packaged as Linux containers.'
	}
]

export default function FaqRoute () {
	const [originalData, setOriginalData] = useState([])
	const [faq, setFaq] = useState([])

	useEffect(() => {
		setFaq(faqArray)
		setOriginalData(faqArray)
	}, [])

	const searchQuestion = (query) => {
		if (!query) return setFaq(originalData)
		if (query === null) return ''

		const res = search(query || '', originalData, { keySelector: (elm) => elm.question || '' })
		setFaq(res)
	}

	return (
		<div className='faq-route-container'>
			<div className='banner'>
				<h1>Frequently asked questions (FAQ)</h1>

				<div className='search-bar'>
					<Search type='text' onChange={searchQuestion} placeholder='Search' />
				</div>
			</div>

			{faq.map(elm => {
				return (
					<Accordion key={elm.question} question={elm.question} answer={elm.answer} />
				)
			})}

			<Pagination />
		</div>
	)
}
