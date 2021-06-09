import React, { useEffect, useState } from 'react'
import Button from '@components/button'
import Input from '@components/input'
// import Select from '@components/select'
import Toogle from '@components/toggle'
import Modals from 'util/modals'
import Api from 'util/api'

import './index.scss'

export default function HomePage () {
	const [test, setTest] = useState(false)
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
			<Input 
				type='text'
				label='Label'
				placeholder='type here'
				value='asadfds'
				onChange={() => console.log('aa')} 
			/>
			<Toogle label='is public' value={test} onChange={() => setTest(!test)} />
		</div>
	)
}