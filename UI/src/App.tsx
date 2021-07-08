import React from 'react'
import { Switch, Route } from 'react-router-dom'
import HomePageRoute from '@routes/home'
import MenuTop from '@components/menuTopBar'
import MenuSideBar from '@components/menuSideBar'
import { TypeSubItem, TypeItem } from './types/generic'
import Modals from './modals'
import 'styles/app.scss'

const SubItens: TypeSubItem[] = [
	{ name: 'Node'}, 
	{ name: 'Nodepools'}
]

const Itens: TypeItem[] = [
	{
		name: 'Cluster',
		icon: 'icon-helm',
		subItens: SubItens
	},
	{ 
		name: 'Machines', 
		icon: 'icon-node',
		subItens: SubItens
	}
]

export default function App() {
	return (
		<div className='route-container'>
			<MenuTop />
			<MenuSideBar 
				itens={Itens}
			/>
			<div className='route-content'>
				<Switch>
						<Route path='/' component={HomePageRoute} />
s				</Switch>
				<Modals />
				</div>
		</div>
	)
}
