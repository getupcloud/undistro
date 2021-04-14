import React from 'react'
import Hand from 'Assets/images/hand.svg'

import './index.scss'

const KeyBenefits = () => {
	return (
		<div className='benefits-container'>
			<div className='content-container'>
				<div className='content'>
					<h1>What is UnDistro?</h1>
					<p><span>UnDistro</span> is a vanilla, non-opinionated, and open source Kubernetes distribution that helps to spin up, manage, and visualize in a standardized and centralized way, one or more production-ready clusters.</p>
					<p>Choosing to use <span>UnDistro</span> is not a matter of choosing <span>UnDistro</span> over EKS/AKS/GKE. Think of <span>UnDistro</span> as an additional “layer” to help you with day 2 K8s operations, no matter the infrastructure, hosted or self-hosted Kubernetes.</p>

					<h1>Why UnDistro?</h1>
					<p><span>UnDistro</span> was created to solve problems found when you have multiple Kubernetes clusters to manage, probably in different infrastructures:</p>
					<div className='bold-content'>
						<p>Multiple Kubernetes clusters add an extra challenge when applying policies and ensuring that all environments are conformant.</p>
						<p>It can be difficult to identify information regarding security, governance, and cluster health.</p>
						<p>Different infrastructures or cloud providers have different experiences for the deployment and management of Kubernetes.</p>
					</div>
				</div>
				<div className='img-container'>
					<img src={Hand} />
				</div>
			</div>
		</div>
	)
}

export default KeyBenefits
