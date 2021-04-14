import React from 'react'
import Button from 'Components/atoms/button'

import './index.scss'
import { useHistory } from 'react-router'

const content = [
	{
		question: 'Q: What is the difference between UnDistro, EKS, AKS, GKE?',
		answer: 'A: Think of UnDistro as an additional “layer” to help you with day 2 operations, no matter the infrastructure or if it is a hosted or self-hosted Kubernetes. Choosing to use unDistro is not a matter of choosing unDistro over any hosted Kubernetes offer. UnDistro is a software that runs inside your Kubernetes management cluster.'
	},
	{
		question: 'Q: I already use a managed Kubernetes. Why should I use UnDistro?',
		answer: 'A: UnDistro uses the Kubernetes reconciliation process to ensure that your cluster is always in the desired state with respect to Kubernetes version, number of nodes, size of nodes, and firewall policies. Another good reason is that if you have more than one cluster to manage UnDistro provides a centralized way to manage multiple clusters.'
	}
]

const Faq = () => {
	const history = useHistory()

	return (
		<div className='faq-container'>
			<div className='content'>
				<h1>FAQ</h1>
				<div className='flex-content'>
					{content.map(elm => {
						return (
							<div key={elm.question} className='questions'>
								<p className='question'>{elm.question}</p>
								<p className='answer'>{elm.answer}</p>
							</div>
						)
					})}
				</div>
				<Button onClick={() => history.push('/faq')} type='read-more'>Read more</Button>
			</div>
		</div>
	)
}

export default Faq
