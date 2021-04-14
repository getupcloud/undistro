import React from 'react'
import Card from 'Components/atoms/horizontalCards'

import './index.scss'

export default function CommunityRoute () {
	return (
		<div className='community-route-container'>
			<div className='banner'>
				<h1>Community</h1>
			</div>

			<div className='community-content'>
				<h1>UnDistro is an awesome open-source project!</h1>
				<h3>UnDistro automates thousands of Kubernetes clusters across multi-cloud, on-prem and edge with unparalleled resilience.</h3>
			</div>

			<Card onClick={() => window.location.replace('https://github.com/getupio-undistro/undistro/discussions')} title='COMMUNITY' title2='Forum' description='Find answers for your questions. Share that great idea you’ve had! This is the place!' />
			<Card onClick={() => window.location.replace('https://github.com/getupio-undistro/undistro/issues')} title='TRACK' title2='Issues' description='Work on tasks, enhancements or report a bug. Get your hands on!' />
		</div>
	)
}
