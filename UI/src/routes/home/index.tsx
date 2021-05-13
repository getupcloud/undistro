import React, { useEffect, useState } from 'react'
import Button from '@components/button'
import Input from '@components/input'
import Select from '@components/select'
import Toogle from '@components/toggle'
import Modals from 'util/modals'
import Api from 'util/api'

import './index.scss'

export default function HomePage () {
	const [test, setTest] = useState(false)
	const [node, setNode] = useState<[]>([])
	const showModal = () => {
    Modals.show('create-cluster', {
      title: 'Create',
			ndTitle: 'Cluster',
			width: '600',
      height: '420'
    })
  }

	const getNodes = () => {
		Api.Node.list()
			.then(res => {
				setNode(res)
			})
	}

	useEffect(() => {
		getNodes()
	}, [])

	console.log(node)

	return (
		<div className='home-page-route'>
			<Button onClick={() => showModal()} size='large' type='primary' children='LgBtnText' />
			<Button size='medium' type='primary' children='LgBtnText' />
			<Button size='large' type='gray' children='LgBtnText' />
			<Button size='medium' type='gray' children='LgBtnText' />
			<Button size='large' type='black' children='LgBtnText' />
			<Button size='medium' type='black' children='LgBtnText' />
			<Input 
				type='text'
				label='Label'
				placeholder='type here'
				value='asadfds'
				onChange={() => console.log('aa')} 
			/>
			<Select label='Label' />
			<Toogle label='is public' value={test} onChange={() => setTest(!test)} />
		</div>
	)
}