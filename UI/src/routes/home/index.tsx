import React, { useEffect, useState } from 'react'
import Button from '@components/button'
import Table from '@components/table'
import Modals from 'util/modals'
import Api from 'util/api'
import moment from 'moment'

import './index.scss'

const headers = [
	{ name: 'Name', field: 'name'},
	{ name: 'Provider', field: 'provider'},
	{ name: 'Flavor', field: 'flavor'},
	{ name: 'Version', field: 'version'},
	{ name: 'Cluster Groups', field: 'clusterGroups'},
	{ name: 'Machines', field: 'machines'},
	{ name: 'Age', field: 'age'},
	{ name: 'Status', field: 'status'},
]

type TypeCluster = {
	name: string,
	provider: string,
	flavor: string,
	version: string,
	clusterGroups: string,
	machines: number,
	age: Date,
	status: string
}

export default function HomePage () {
	const [clusters, setClusters] = useState<TypeCluster[]>([])
	const [pause, setPause] = useState<boolean>(false)
	const name = (clusters || []).map(elm => elm.name).toString()
	const namespace = (clusters || []).map(elm => elm.clusterGroups).toString()


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
						clusterGroups: elm.metadata.namespace,
						machines: elm.status.controlPlane.replicas + elm.status.totalWorkerReplicas,
						age: moment(elm.metadata.creationTimestamp).startOf('day').fromNow(),
						status: elm.status.conditions[0].type
					}
				}))
			})
	}

	const pauseCluster = () => {
		setPause(!pause)	
		const payload = {
			"spec": {
				"paused": pause
			}
		}

		Api.Cluster.put(payload, namespace, name)
			.then(_ => {
				console.log('success')
			})
	}

	const deleteCluster = () => {
		Api.Cluster.delete(namespace, name)
			.then(_ => {
				console.log('success')
			})
	}


	useEffect(() => {
		getClusters()
	}, [])

	console.log(pause)

	return (
		<div className='home-page-route'>
			<Button onClick={() => showModal()} size='large' type='primary' children='LgBtnText' />
			<Table data={(clusters || [])} icon={pause} delete={() => deleteCluster()} pause={() => pauseCluster()} header={headers}/>	
		</div>
	)
}