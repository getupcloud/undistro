import React, { FC } from 'react'
import Item from '@components/itemSideMenu'
import './index.scss'

const MenuSideBar: FC = () => {

	return (
		<div className='menu-side-container'>
			<ul className='side-itens'>
				<Item name='Cluster' icon='icon-cluster' />
				<Item name='Nodes' icon='icon-cluster' />
				<Item name='Network' icon='icon-network' />
				<Item name='Configuration' icon='icon-workload' />
				<Item name='Item 05' icon='icon-cluster' />
				<Item name='Item 06' icon='icon-network' />
				<Item name='Item 07' icon='icon-workload' />
				<Item name='Item 08' icon='icon-cluster' />
			</ul>			
		</div>
	)
}

export default MenuSideBar