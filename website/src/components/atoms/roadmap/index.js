import React from 'react'

import './index.scss'

const Roadmap = () => {
	return (
		<div className='roadmap-container'>
			<div className='content'>
  			<h1><span>UnDistro</span> RoadMap</h1>
				<div className='cards'>
					<div className='card'>
						<div className='title'>
							<i className='icon-check' />
							<h2>Phase 1:</h2>
						</div>
						<p>Cluster lifecycle, Helm Charts management. Infrastructure supported: AWS</p>
					</div>
					<div className='card'>
						<div className='title'>
							<i className='icon-working' />
							<h2>Phase 2:</h2>
						</div>
						<p>Web UI, RBAC management. Infrastructure supported: Azure</p>
					</div>
					<div className='card'>
						<div className='title'>
							<i className='icon-scheduled' />
							<h2>Phase 3:</h2>
						</div>
						<p>Code to production. Infrastructure supported: On Prem, Google Cloud, VMWare, OpenStack</p>
					</div>
					<div className='card'>
						<div className='title'>
							<i className='icon-scheduled' />
							<h2>Phase 4:</h2>
						</div>
						<p>OPA, Telemetry, Service Mesh, Network Policies</p>
					</div>
				</div>
			</div>
		</div>
	)
}

export default Roadmap
