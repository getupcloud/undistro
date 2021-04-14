import React from 'react'
import Button from 'Components/atoms/button'
import Conections from 'Assets/images/img-highlight.svg'
import StarConection from 'Assets/images/img2-highlight.svg'
import { useHistory } from 'react-router'

import './index.scss'

const Highlight = () => {
	const history = useHistory()

	return (
		<div className='highlight-container'>
			<div className='float-title'>
				<p>Highlight</p>
			</div>
			<div className='content'>
				<div className='text-and-button'>
					<h1>Different infrastructures, same experience.</h1>
					<p>UnDistro uses <span onClick={() => window.open('https://cluster-api.sigs.k8s.io/')}>Cluster API</span> to work as an additional layer, abstracting the particularities of each infrastructure.</p>
  				<Button onClick={() => history.push('/docs#2---introduction')} >More</Button>
				</div>
				<div className='text-and-img'>
					<div className='img-container'>
						<img className='connections' src={Conections} />
					</div>
					<div>
						<h1>Actionable information</h1>
						<p>A control plane where you can find health reports across multiple Kubernetes clusters easily  provided by Prometheus federated mode.</p>
						<span>(under development, see roadmap)</span>
					</div>
				</div>
				<div className='text-and-img'>
					<div className='img-container'>
						<img className='star' src={StarConection} />
					</div>
					<div>
						<h1>Centralized cluster management</h1>
						<p>UnDistro creates a management cluster giving you a standardized and centralized way to manage multiple Kubernetes clusters.</p>
					</div>
				</div>
			</div>
		</div>
	)
}

export default Highlight
