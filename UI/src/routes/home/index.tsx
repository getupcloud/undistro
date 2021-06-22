import React from 'react'
import Button from '@components/button'
import Table from '@components/table'
import Modals from 'util/modals'
import './index.scss'

export default function HomePage () {
	const showModal = () => {
    Modals.show('create-cluster', {
      title: 'Create',
			ndTitle: 'Cluster',
			width: '600',
      height: '420'
    })
  }

	return (
		<div className='home-page-route'>
			<Button onClick={() => showModal()} size='large' type='primary' children='LgBtnText' />
			<Table />
		</div>
	)
}