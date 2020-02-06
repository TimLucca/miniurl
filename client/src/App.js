import React from 'react';
import './App.css';
import {BrowserRouter as Router, Route, Switch} from 'react-router-dom'
import {Container} from 'semantic-ui-react'
import Home from './home/Home'
import Nav from './components/Nav'
import Stats from './components/Stats'

const App = () => {
  return (
      <Router>
        <div>
          <Nav />
          <Container>
            <Switch>
              <Route exact path='/' component={Home}/>
              <Route exact path='/stats' component={Stats} />
            </Switch>
          </Container>
        </div>
      </Router>
  );
}

export default App;
