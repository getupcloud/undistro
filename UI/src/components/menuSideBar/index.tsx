import React, { FC, useState } from 'react'
import Classnames from 'classnames'
import { TypeMenu } from '../../types/generic'
import './index.scss'

const MenuSideBar: FC<TypeMenu> = ({ 
	itens
}) => {
	const [show, setShow] = useState<boolean>(false)
	const style = Classnames('side-item', {
		active: show
	})

	return (
		<div className='menu-side-container'>
			<ul className='side-itens'>
				{itens.map((elm: any) => {
					return (
						<>
							<li onClick={() => setShow(!show)} className={style}>
								<i className={elm.icon} />
								<p>{elm.name}</p>
								{show ? <i className='icon-arrow-up' /> : <i className='icon-arrow-down' />}
							</li>
							{show && <div className='item-menu'>
								{elm.subItens.map((elm: Partial<{name: string}>) => (<p>{elm.name}</p>))}
							</div>}
						</>
					)
				})}
			</ul>			
		</div>
	)
}

export default MenuSideBar