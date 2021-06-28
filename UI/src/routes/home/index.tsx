import React, { useEffect, useState } from 'react'
import Button from '@components/button'
import Table from '@components/table'
import Modals from 'util/modals'
import Api from 'util/api'
import moment from 'moment'

import './index.scss'

export default function HomePage () {
	const [clusters, setClusters] = useState()
	const showModal = () => {
    Modals.show('create-cluster', {
      title: 'Create',
			ndTitle: 'Cluster',
			width: '600',
      height: '420'
    })
  }

	moment.updateLocale('en', {
    relativeTime : {
        past:   "%s",
        s  : 's',
        ss : '%ds',
        m:  "m",
        mm: "%dm",
        h:  "h",
        hh: "%dh",
        d:  "d",
        dd: "%dd",
        M:  "m",
        MM: "%dm",
        y:  "y",
        yy: "%dy"
    }
})

	const getClusters = () => {
		Api.Cluster.list('undistro-system')
			.then((clusters) => {
				setClusters(clusters.items.map((elm: any) => {
					return {
						name: elm.metadata.name,
						provider: elm.spec.infrastructureProvider.name,
						flavor: elm.spec.infrastructureProvider.flavor,
						version: elm.spec.kubernetesVersion,
						age: moment(elm.metadata.creationTimestamp).startOf('day').fromNow(),
						status: elm.status.conditions[0].type
					}
				}))
			})
	}


	const headers = [
		{ name: 'Name', field: 'name'},
		{ name: 'Provider', field: 'provider'},
		{ name: 'Flavor', field: 'flavor'},
		{ name: 'Version', field: 'version'},
		// { name: 'IP address', field: 'ver'},
		{ name: 'Age', field: 'age'},
		{ name: 'Status', field: 'status'},
	]

	console.log(clusters)
	useEffect(() => {
		getClusters()
	}, [])

	return (
		<div className='home-page-route'>
			<ul>
				<li><i className='icon-stop' /> <p>Stop</p></li>
				<li><i className='icon-arrow-solid-up' /> <p>Update</p></li>
				<li><i className='icon-settings' /> <p>Settings</p></li>
				<li><i className='icon-close-solid' /> <p>Delete</p></li>
			</ul>
			<Button onClick={() => showModal()} size='large' type='primary' children='LgBtnText' />
			<Table data={(clusters || [])}  header={headers}/>	
		</div>
	)
}