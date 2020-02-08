import React, {Component} from 'react'
import axios from "axios"
import { Header, Form, Input } from 'semantic-ui-react'

class Home extends Component {
  constructor(props) {
    super(props)

    this.state = {
      long: '',
      res: '',
      submitted: false,
    }
  }

  componentDidMount() {
    
  }

  onChange = event => {
    this.setState({
      [event.target.name]: event.target.value
    })
  }

  onSubmit = () => {
    let {long} = this.state 

    if (long) {
      axios.post(
        '/api/new',
        {long},
        {
          headers: {
            'Content-Type': 'application/json'
          }
        }
      ).then(res => {
        console.log(res)
        this.setState({res: res, submitted: true})
      })
    }
  }

  isMini = () => {
    let res = this.state.res
    console.log(this.state)
    if(!this.state.submitted){
      return
    }
    if (res.status === 200) {
      return (
        <p><strong>MiniURL: </strong><a href={res.data.miniurl}>{res.data.miniurl}</a></p>
      )
    } else if(!this.state.submitted){
      return
    } else {
      return (
        <div className='ui negative message'>
          <i className='close icon'></i>
          <div className='header'>
            An error occured
          </div>
        </div>
      )
    }
  }

  render() {
    const styles = {
      paddingBottom: '2em',
    }

    return (    
      <div>
        <div className='row' style={styles}>
          <Header className='header' as='h1'>
            MiniURL
          </Header>
        </div>
        <div className='row' style={styles}>
          <Form onSubmit={this.onSubmit}>
            <Input type='text' name='long' value={this.state.long} onChange={this.onChange} fluid placeholder='Enter URL' />
          </Form>
        </div>
        <button onClick={this.onSubmit}>Minify</button>
        {this.isMini()}
      </div>
    )
  }
}

export default Home;