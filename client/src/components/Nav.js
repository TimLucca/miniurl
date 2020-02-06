import React, { Component } from 'react'
import { Menu } from 'semantic-ui-react'
import {Link} from 'react-router-dom'

export default class Nav extends Component {
  state = {}

  render() {
    const { activeItem } = this.state

    return (
      <Menu>
        <Menu.Item
          name='Home'
          active={activeItem === 'home'}
        >
          <Link to='/'>Home</Link>
        </Menu.Item>

        <Menu.Item
          name='Statistics'
          active={activeItem === 'stats'}
        >
          <Link to='/stats'>Statistics</Link>
        </Menu.Item>
      </Menu>
    )
  }
}
